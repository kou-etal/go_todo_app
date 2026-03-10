package refresh

import (
	"context"
)

type RefreshTokenRepository interface {
	Store(ctx context.Context, t *RefreshToken) error
	FindByTokenHashForUpdate(ctx context.Context, hash TokenHash) (*RefreshToken, error)
	Update(ctx context.Context, t *RefreshToken) error
}
