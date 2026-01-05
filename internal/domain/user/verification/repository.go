package verification

import (
	"context"
)

type EmailVerificationRepository interface {
	Insert(ctx context.Context, t *EmailVerificationToken) error
}
