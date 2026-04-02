package login

import (
	"context"
	"errors"
	"time"

	"github.com/kou-etal/go_todo_app/internal/clock"
	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	drefresh "github.com/kou-etal/go_todo_app/internal/domain/user/refresh"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("usecase/user/login")

type AccessTokenGenerator interface {
	GenerateAccessToken(userID string) (string, error)
}

type Usecase struct {
	tx             usetx.Runner[usetx.LoginDeps]
	clock          clock.Clocker
	passwordHasher duser.PasswordHasher
	tokenGen       AccessTokenGenerator
	refreshHasher  drefresh.TokenHasher
	refreshGen     drefresh.TokenGenerator
	refreshTTL     time.Duration
	accessTTLSec   int
}

func New(
	tx usetx.Runner[usetx.LoginDeps],
	clock clock.Clocker,
	passwordHasher duser.PasswordHasher,
	tokenGen AccessTokenGenerator,
	refreshHasher drefresh.TokenHasher,
	refreshGen drefresh.TokenGenerator,
	refreshTTL time.Duration,
	accessTTLSec int,
) *Usecase {
	return &Usecase{
		tx:             tx,
		clock:          clock,
		passwordHasher: passwordHasher,
		tokenGen:       tokenGen,
		refreshHasher:  refreshHasher,
		refreshGen:     refreshGen,
		refreshTTL:     refreshTTL,
		accessTTLSec:   accessTTLSec,
	}
}

func (u *Usecase) Do(ctx context.Context, cmd Command) (Result, error) {
	ctx, span := tracer.Start(ctx, "user.login")
	defer span.End()

	cmd, err := normalize(cmd)
	if err != nil {
		return Result{}, err
	}

	email, err := duser.NewUserEmail(cmd.Email)
	if err != nil {
		return Result{}, ErrInvalidCredentials
	}

	now := u.clock.Now()

	var accessToken string
	var plainRefresh string

	if err := u.tx.WithinTx(ctx, func(ctx context.Context, deps usetx.LoginDeps) error {
		usr, err := deps.UserRepo().FindByEmail(ctx, email)
		if err != nil {
			if errors.Is(err, duser.ErrNotFound) {
				return ErrInvalidCredentials
			}
			return err
		}

		if err := usr.Password().Compare(cmd.Password, u.passwordHasher); err != nil {
			return ErrInvalidCredentials
		}

		accessToken, err = u.tokenGen.GenerateAccessToken(usr.ID().Value())
		if err != nil {
			return err
		}

		rt, plain, err := drefresh.NewRefreshToken(
			usr.ID(),
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
		if errors.Is(err, ErrInvalidCredentials) {
			return Result{}, ErrInvalidCredentials
		}
		return Result{}, err
	}

	return Result{
		AccessToken:  accessToken,
		RefreshToken: plainRefresh,
		ExpiresIn:    u.accessTTLSec,
	}, nil
}
