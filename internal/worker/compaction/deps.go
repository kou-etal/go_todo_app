package compaction

import (
	"context"
	"io"
)

// upload抽象
// ObjectStorage は compaction ワーカーが必要とする S3 操作の契約。
// outbox の ObjectUploader より広い（List, Get が追加）。
type ObjectStorage interface {
	Upload(ctx context.Context, key string, body io.Reader) error
	Exists(ctx context.Context, key string) (bool, error)
	List(ctx context.Context, prefix string) ([]string, error)
	Get(ctx context.Context, key string) (io.ReadCloser, error)
}
