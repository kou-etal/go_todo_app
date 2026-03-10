package txdeps

import (
	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	drefresh "github.com/kou-etal/go_todo_app/internal/domain/user/refresh"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

type logindeps struct {
	userRepo         duser.UserRepository
	refreshTokenRepo drefresh.RefreshTokenRepository
}

var _ usetx.LoginDeps = (*logindeps)(nil)

func NewLogin(
	u duser.UserRepository,
	r drefresh.RefreshTokenRepository,
) usetx.LoginDeps {
	return &logindeps{
		userRepo:         u,
		refreshTokenRepo: r,
	}
}

func (d *logindeps) UserRepo() duser.UserRepository {
	return d.userRepo
}

func (d *logindeps) RefreshTokenRepo() drefresh.RefreshTokenRepository {
	return d.refreshTokenRepo
}
