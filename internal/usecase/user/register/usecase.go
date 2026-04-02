package register

import (
	"context"
	"errors"

	"github.com/kou-etal/go_todo_app/internal/clock"
	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("usecase/user/register")

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
	ctx, span := tracer.Start(ctx, "user.register")
	defer span.End()

	cmd, err := normalize(cmd)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return Result{}, err
	}

	now := u.clock.Now()

	email, err := duser.NewUserEmail(cmd.Email)
	if err != nil {
		switch {
		case errors.Is(err, duser.ErrEmailTooLong):
			span.RecordError(ErrEmailTooLong)
			span.SetStatus(codes.Error, ErrEmailTooLong.Error())
			return Result{}, ErrEmailTooLong
		case errors.Is(err, duser.ErrInvalidEmailFormat):
			span.RecordError(ErrInvalidEmailFormat)
			span.SetStatus(codes.Error, ErrInvalidEmailFormat.Error())
			return Result{}, ErrInvalidEmailFormat
		default:
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return Result{}, err
		}
	}

	pass, err := duser.NewUserPasswordFromPlain(cmd.Password, u.passwordHasher)
	if err != nil {
		switch {
		case errors.Is(err, duser.ErrPasswordTooShort):
			span.RecordError(ErrPasswordTooShort)
			span.SetStatus(codes.Error, ErrPasswordTooShort.Error())
			return Result{}, ErrPasswordTooShort
		case errors.Is(err, duser.ErrPasswordTooLong):
			span.RecordError(ErrPasswordTooLong)
			span.SetStatus(codes.Error, ErrPasswordTooLong.Error())
			return Result{}, ErrPasswordTooLong
		case errors.Is(err, duser.ErrPasswordHasLeadingOrTrailingSpace):
			span.RecordError(ErrPasswordHasLeadingOrTrailingSpace)
			span.SetStatus(codes.Error, ErrPasswordHasLeadingOrTrailingSpace.Error())
			return Result{}, ErrPasswordHasLeadingOrTrailingSpace
		default:
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return Result{}, err
		}
	}

	name, err := duser.NewUserName(cmd.UserName)
	if err != nil {
		switch {
		case errors.Is(err, duser.ErrNameTooLong):
			span.RecordError(ErrNameTooLong)
			span.SetStatus(codes.Error, ErrNameTooLong.Error())
			return Result{}, ErrNameTooLong
		default:
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return Result{}, err
		}
	}

	user := duser.NewUser(email, pass, name, now)

	span.SetAttributes(attribute.String("user.id", user.ID().Value()))

	token, plain, err := dverify.NewEmailVerificationToken(
		user.ID(),
		u.tokenHasher,
		u.tokenGenerator,
		now,
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
			span.RecordError(ErrConflict)
			span.SetStatus(codes.Error, ErrConflict.Error())
			return Result{}, ErrConflict
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return Result{}, err
	}

	return Result{UserID: user.ID().Value()}, nil
}
