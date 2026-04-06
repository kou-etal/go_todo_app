package txdeps

import (
	drefresh "github.com/kou-etal/go_todo_app/internal/domain/user/refresh"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

type refreshdeps struct {
	refreshTokenRepo drefresh.RefreshTokenRepository
}

var _ usetx.RefreshDeps = (*refreshdeps)(nil)

func NewRefresh(
	r drefresh.RefreshTokenRepository,
) usetx.RefreshDeps {
	return &refreshdeps{
		refreshTokenRepo: r,
	}
}

func (d *refreshdeps) RefreshTokenRepo() drefresh.RefreshTokenRepository {
	return d.refreshTokenRepo
}
