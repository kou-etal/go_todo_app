package tx

import "context"

type Runner[D any] interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, deps D) error) error
}
