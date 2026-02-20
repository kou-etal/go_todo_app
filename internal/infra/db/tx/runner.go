package txrunner

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kou-etal/go_todo_app/internal/infra/db"
)

type DepsFactory[D any] func(q db.QueryerExecer) D

type SQLxRunner[D any] struct {
	beginner   db.Beginner
	opts       *sql.TxOptions
	makeTxDeps DepsFactory[D]
}

func New[D any](beginner db.Beginner, opts *sql.TxOptions, makeTxDeps DepsFactory[D]) *SQLxRunner[D] {
	return &SQLxRunner[D]{
		beginner:   beginner,
		opts:       opts,
		makeTxDeps: makeTxDeps,
	}
}

func (r *SQLxRunner[D]) WithinTx(
	ctx context.Context,
	fn func(ctx context.Context, deps D) error,
) (retErr error) {
	//deferで追記するから名前付き戻り値
	tx, err := r.beginner.BeginTxx(ctx, r.opts)
	if err != nil {
		return err
	}

	committed := false

	defer func() {

		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}

		if retErr != nil && !committed {

			if rbErr := tx.Rollback(); rbErr != nil {
				retErr = fmt.Errorf("tx rollback failed: %v (original: %w)", rbErr, retErr)
			}
		}
	}()

	deps := r.makeTxDeps(tx)

	if err := fn(ctx, deps); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {

		committed = true
		return err
	}
	committed = true
	return nil
}
