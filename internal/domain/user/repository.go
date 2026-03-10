package user

import (
	"context"
)

type UserRepository interface {
	Store(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email UserEmail) (*User, error)
}
