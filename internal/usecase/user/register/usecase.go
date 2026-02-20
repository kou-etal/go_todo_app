package register

import (
	"context"
	"errors"

	"github.com/kou-etal/go_todo_app/internal/clock"
	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

type Usecase struct {
	tx             usetx.Runner[usetx.RegisterDeps]
	clock          clock.Clocker
	passwordHasher duser.PasswordHasher
	tokenGenerator dverify.TokenGenerator
	tokenHasher    dverify.TokenHasher
}

func New(
	tx usetx.Runner[usetx.RegisterDeps],
	clock clock.Clocker,
	passwordHasher duser.PasswordHasher,
	tokenGenerator dverify.TokenGenerator,
	tokenHasher dverify.TokenHasher,
) *Usecase {
	return &Usecase{
		tx:             tx,
		clock:          clock,
		passwordHasher: passwordHasher,
		tokenGenerator: tokenGenerator,
		tokenHasher:    tokenHasher,
	}
}

func (u *Usecase) Do(ctx context.Context, cmd Command) (Result, error) {
	cmd, err := normalize(cmd)
	if err != nil {
		return Result{}, err
	}

	now := u.clock.Now()

	email, err := duser.NewUserEmail(cmd.Email)
	if err != nil {
		switch {
		case errors.Is(err, duser.ErrEmailTooLong):
			return Result{}, ErrEmailTooLong
		case errors.Is(err, duser.ErrInvalidEmailFormat):
			return Result{}, ErrInvalidEmailFormat
		default:
			return Result{}, err
		}
	}

	pass, err := duser.NewUserPasswordFromPlain(cmd.Password, u.passwordHasher)
	if err != nil {
		switch {
		case errors.Is(err, duser.ErrPasswordTooShort):
			return Result{}, ErrPasswordTooShort
		case errors.Is(err, duser.ErrPasswordTooLong):
			return Result{}, ErrPasswordTooLong
		case errors.Is(err, duser.ErrPasswordHasLeadingOrTrailingSpace):
			return Result{}, ErrPasswordHasLeadingOrTrailingSpace
		default:
			return Result{}, err
		}
	}

	name, err := duser.NewUserName(cmd.UserName)
	if err != nil {
		switch {
		case errors.Is(err, duser.ErrNameTooLong):
			return Result{}, ErrNameTooLong
		default:
			return Result{}, err
		}
	}

	user := duser.NewUser(email, pass, name, now)

	token, plain, err := dverify.NewEmailVerificationToken(
		user.ID(),
		u.tokenHasher,
		u.tokenGenerator,
		now,
	)
	if err != nil {
		return Result{}, err
	}

	err = u.tx.WithinTx(ctx, func(ctx context.Context, deps usetx.RegisterDeps) error {
		if err := deps.UserRepo().Store(ctx, user); err != nil {
			return err
		}
		if err := deps.EmailVerifyRepo().Insert(ctx, token); err != nil {
			return err
		}

		_ = plain //TODO:mailer

		return nil
	})
	if err != nil {
		if errors.Is(err, duser.ErrConflict) {
			return Result{}, ErrConflict
		}
		return Result{}, err
	}

	return Result{UserID: user.ID().Value()}, nil
}
