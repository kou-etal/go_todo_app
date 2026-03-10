package refresh

import (
	"context"
	"errors"
	"time"

	"github.com/kou-etal/go_todo_app/internal/clock"
	drefresh "github.com/kou-etal/go_todo_app/internal/domain/user/refresh"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

type AccessTokenGenerator interface {
	GenerateAccessToken(userID string) (string, error)
}

type Usecase struct {
	tx           usetx.Runner[usetx.RefreshDeps]
	clock        clock.Clocker
	tokenGen     AccessTokenGenerator
	refreshHasher drefresh.TokenHasher
	refreshGen   drefresh.TokenGenerator
	refreshTTL   time.Duration
	accessTTLSec int
}

func New(
	tx usetx.Runner[usetx.RefreshDeps],
	clock clock.Clocker,
	tokenGen AccessTokenGenerator,
	refreshHasher drefresh.TokenHasher,
	refreshGen drefresh.TokenGenerator,
	refreshTTL time.Duration,
	accessTTLSec int,
) *Usecase {
	return &Usecase{
		tx:            tx,
		clock:         clock,
		tokenGen:      tokenGen,
		refreshHasher: refreshHasher,
		refreshGen:    refreshGen,
		refreshTTL:    refreshTTL,
		accessTTLSec:  accessTTLSec,
	}
}

func (u *Usecase) Do(ctx context.Context, cmd Command) (Result, error) {
	cmd, err := normalize(cmd)
	if err != nil {
		return Result{}, err
	}

	hash, err := drefresh.NewTokenHashFromPlain(cmd.RefreshToken, u.refreshHasher)
	if err != nil {
		return Result{}, ErrInvalidRefreshToken
	}

	now := u.clock.Now()

	var accessToken string
	var plainRefresh string

	if err := u.tx.WithinTx(ctx, func(ctx context.Context, deps usetx.RefreshDeps) error {
		old, err := deps.RefreshTokenRepo().FindByTokenHashForUpdate(ctx, hash)
		if err != nil {
			if errors.Is(err, drefresh.ErrNotFound) {
				return ErrInvalidRefreshToken
			}
			return err
		}

		if !old.IsValid(now) {
			return ErrInvalidRefreshToken
		}

		old.Revoke(now)
		if err := deps.RefreshTokenRepo().Update(ctx, old); err != nil {
			return err
		}

		accessToken, err = u.tokenGen.GenerateAccessToken(old.UserID().Value())
		if err != nil {
			return err
		}

		rt, plain, err := drefresh.NewRefreshToken(
			old.UserID(),
			u.refreshHasher,
			u.refreshGen,
			u.refreshTTL,
			now,
		)
		if err != nil {
			return err
		}
		plainRefresh = plain

		return deps.RefreshTokenRepo().Store(ctx, rt)
	}); err != nil {
		if errors.Is(err, ErrInvalidRefreshToken) {
			return Result{}, ErrInvalidRefreshToken
		}
		return Result{}, err
	}

	return Result{
		AccessToken:  accessToken,
		RefreshToken: plainRefresh,
		ExpiresIn:    u.accessTTLSec,
	}, nil
}
