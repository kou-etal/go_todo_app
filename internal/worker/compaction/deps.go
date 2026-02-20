package compaction

import (
	"context"
	"io"
)

type ObjectStorage interface {
	Upload(ctx context.Context, key string, body io.Reader) error
	Exists(ctx context.Context, key string) (bool, error)
	List(ctx context.Context, prefix string) ([]string, error)
	Get(ctx context.Context, key string) (io.ReadCloser, error)
}
