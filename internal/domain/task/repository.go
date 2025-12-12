package task

import "context"

//type Repositoryの議論
type TaskRepository interface {
	//Save(ctx context.Context, t *Task) error
	ListAll(ctx context.Context) ([]*Task, error)
}
