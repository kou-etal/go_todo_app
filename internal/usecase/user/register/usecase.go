package register

//withintxはusecaseで使わずにappで作るのが良い。txをusecaseで作ると厚くなる。
//それに伴ってそもそもtxrunnerをstructのフィールドに入れない。usecase実装からtxを排除
//それしたらusecaseがいつも通りになって見やすいな。

import (
	"context"

	"github.com/kou-etal/go_todo_app/internal/clock"
	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

type Usecase struct {
	tx             usetx.Runner
	clock          clock.Clocker
	passwordHasher duser.PasswordHasher
	tokenGenerator dverify.TokenGenerator
	tokenHasher    dverify.TokenHasher
}

func New(
	tx usetx.Runner,
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
		return Result{}, err
	}

	pass, err := duser.NewUserPasswordFromPlain(cmd.Password, u.passwordHasher)
	if err != nil {
		return Result{}, err
	}

	name, err := duser.NewUserName(cmd.UserName)
	if err != nil {
		return Result{}, err
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

	var res Result

	err = u.tx.WithinTx(ctx, func(ctx context.Context, deps usetx.RegisterDeps) error {
		if err := deps.UserRepo().Store(ctx, user); err != nil {
			return err
		}
		if err := deps.EmailVerifyRepo().Insert(ctx, token); err != nil {
			return err
		}

		_ = plain //TODO:mailer
		res = Result{UserID: user.ID().Value()}
		return nil
	})
	if err != nil {
		return Result{}, err
	}

	return res, nil
}

//whithintxはappで組み立てるから使わなくなったコード
/*err = u.txRunner.WithinTx(ctx, func(ctx context.Context, deps usetx.Deps) error {
	if err := deps.Users.Store(ctx, user); err != nil {
		return err
	}
	if err := deps.EmailTokens.Insert(ctx, token); err != nil {
		return err
	}
	return nil
})*/
