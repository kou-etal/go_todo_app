package txdeps

import (
	dtaskevent "github.com/kou-etal/go_todo_app/internal/domain/event"
	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

type taskeventdeps struct {
	taskRepo      dtask.TaskRepository
	taskEventRepo dtaskevent.TaskEventRepository
}

var _ usetx.TaskEventDeps = (*taskeventdeps)(nil)

func NewTaskEvent(
	t dtask.TaskRepository,
	e dtaskevent.TaskEventRepository,
) usetx.TaskEventDeps {
	return &taskeventdeps{
		taskRepo:      t,
		taskEventRepo: e,
	}
}

func (d *taskeventdeps) TaskRepo() dtask.TaskRepository {
	return d.taskRepo
}

func (d *taskeventdeps) TaskEventRepo() dtaskevent.TaskEventRepository {
	return d.taskEventRepo
}
