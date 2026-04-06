package task

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/kou-etal/go_todo_app/internal/auth"
	"github.com/kou-etal/go_todo_app/internal/domain/user"
	"github.com/kou-etal/go_todo_app/internal/logger"

	taskevent "github.com/kou-etal/go_todo_app/internal/domain/event"
	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

const testUserID = "00000000-0000-0000-0000-000000000001"

func withAuthContext(r *http.Request) *http.Request {
	ctx := auth.WithUserID(r.Context(), user.UserID(testUserID))
	return r.WithContext(ctx)
}

var errInternal = errors.New("internal")

// --- stub logger (no-op) ---

type stubLogger struct{}

func (stubLogger) Debug(_ context.Context, _ string, _ error, _ ...logger.Attr) {}
func (stubLogger) Info(_ context.Context, _ string, _ ...logger.Attr)            {}
func (stubLogger) Warn(_ context.Context, _ string, _ ...logger.Attr)            {}
func (stubLogger) Error(_ context.Context, _ string, _ error, _ ...logger.Attr)  {}

// --- stub clocker ---

type stubClocker struct{ now time.Time }

func (c stubClocker) Now() time.Time { return c.now }

// --- mock runner for TaskEventDeps ---

type mockTaskEventRunner struct {
	deps    usetx.TaskEventDeps
	withErr error
}

func (m *mockTaskEventRunner) WithinTx(ctx context.Context, fn func(ctx context.Context, deps usetx.TaskEventDeps) error) error {
	if m.withErr != nil {
		return m.withErr
	}
	return fn(ctx, m.deps)
}

// --- mock task event deps ---

type mockTaskEventDeps struct {
	taskRepo      dtask.TaskRepository
	taskEventRepo taskevent.TaskEventRepository
}

func (m *mockTaskEventDeps) TaskRepo() dtask.TaskRepository              { return m.taskRepo }
func (m *mockTaskEventDeps) TaskEventRepo() taskevent.TaskEventRepository { return m.taskEventRepo }

// --- mock task repo ---

type mockTaskRepo struct {
	storeErr  error
	updateErr error
	deleteErr error
	findTask  *dtask.Task
	findErr   error
	listTasks []*dtask.Task
	listNext  *dtask.ListCursor
	listErr   error
}

func (m *mockTaskRepo) Store(_ context.Context, _ *dtask.Task) error { return m.storeErr }
func (m *mockTaskRepo) Update(_ context.Context, _ *dtask.Task) error { return m.updateErr }
func (m *mockTaskRepo) Delete(_ context.Context, _ dtask.TaskID, _ uint64) error {
	return m.deleteErr
}
func (m *mockTaskRepo) FindByID(_ context.Context, _ dtask.TaskID) (*dtask.Task, error) {
	return m.findTask, m.findErr
}
func (m *mockTaskRepo) List(_ context.Context, _ dtask.ListQuery) ([]*dtask.Task, *dtask.ListCursor, error) {
	return m.listTasks, m.listNext, m.listErr
}

// --- mock event repo ---

type mockEventRepo struct {
	insertErr error
}

func (m *mockEventRepo) Insert(_ context.Context, _ *taskevent.TaskEvent) error {
	return m.insertErr
}
