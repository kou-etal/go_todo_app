package taskevent

import (
	"context"
)

type TaskEventRepository interface {
	Insert(ctx context.Context, e *TaskEvent) error
}
