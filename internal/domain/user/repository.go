package user

import (
	"context"
)

type UserRepository interface {
	Store(ctx context.Context, u *User) error
}
