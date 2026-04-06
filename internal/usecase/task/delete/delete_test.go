package remove

import (
	"context"
	"errors"
	"testing"
	"time"

	taskevent "github.com/kou-etal/go_todo_app/internal/domain/event"
	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	"github.com/kou-etal/go_todo_app/internal/domain/user"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

// --- mock runner ---

type mockRunner struct {
	deps    usetx.TaskEventDeps
	withErr error // WithinTx自体が返すエラー（fn内エラーではなくrunner側エラー）
}

func (m *mockRunner) WithinTx(ctx context.Context, fn func(ctx context.Context, deps usetx.TaskEventDeps) error) error {
	if m.withErr != nil {
		return m.withErr
	}
	return fn(ctx, m.deps)
}

// --- mock deps ---

type mockDeps struct {
	taskRepo      dtask.TaskRepository
	taskEventRepo taskevent.TaskEventRepository
}

func (m *mockDeps) TaskRepo() dtask.TaskRepository              { return m.taskRepo }
func (m *mockDeps) TaskEventRepo() taskevent.TaskEventRepository { return m.taskEventRepo }

// --- mock task repo ---

type mockTaskRepo struct {
	findTask *dtask.Task
	findErr  error

	deleteErr error

	// 記録用
	deleteCalled bool
	deleteID     dtask.TaskID
}

func (m *mockTaskRepo) FindByID(_ context.Context, id dtask.TaskID) (*dtask.Task, error) {
	return m.findTask, m.findErr
}

func (m *mockTaskRepo) Delete(_ context.Context, id dtask.TaskID, version uint64) error {
	m.deleteCalled = true
	m.deleteID = id
	return m.deleteErr
}

func (m *mockTaskRepo) List(_ context.Context, _ dtask.ListQuery) ([]*dtask.Task, *dtask.ListCursor, error) {
	return nil, nil, nil
}
func (m *mockTaskRepo) Store(_ context.Context, _ *dtask.Task) error   { return nil }
func (m *mockTaskRepo) Update(_ context.Context, _ *dtask.Task) error  { return nil }

// --- mock event repo ---

type mockEventRepo struct {
	insertErr    error
	insertCalled bool
	insertedEvent *taskevent.TaskEvent
}

func (m *mockEventRepo) Insert(_ context.Context, e *taskevent.TaskEvent) error {
	m.insertCalled = true
	m.insertedEvent = e
	return m.insertErr
}

// --- mock clocker ---

type mockClocker struct {
	now time.Time
}

func (m mockClocker) Now() time.Time { return m.now }

// --- helper ---

const testUID = "00000000-0000-0000-0000-000000000001"

func testTask(id dtask.TaskID, dueDate time.Time) *dtask.Task {
	title, _ := dtask.NewTaskTitle("test title")
	desc, _ := dtask.NewTaskDescription("test desc")
	due, _ := dtask.NewDueDateFromTime(dueDate)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return dtask.ReconstructTask(id, user.UserID(testUID), title, desc, dtask.StatusTodo, due, now, now, 1)
}

// --- tests ---

func TestDo_happyPath(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC) // 7日後
	taskID := dtask.NewTaskID()

	taskRepo := &mockTaskRepo{findTask: testTask(taskID, dueDate)}
	eventRepo := &mockEventRepo{}
	deps := &mockDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockRunner{deps: deps}
	clk := mockClocker{now: now}

	uc := New(runner, clk)
	err := uc.Do(context.Background(), Command{UserID: testUID, ID: taskID.Value(), Version: 1})
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}

	if !taskRepo.deleteCalled {
		t.Fatal("Delete was not called")
	}
	if !eventRepo.insertCalled {
		t.Fatal("Insert was not called")
	}
	if eventRepo.insertedEvent.EventType() != taskevent.EventDeleted {
		t.Fatalf("event type = %s, want deleted", eventRepo.insertedEvent.EventType())
	}
}

func TestDo_normalizeError_emptyID(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{}
	clk := mockClocker{now: time.Now()}
	uc := New(runner, clk)

	err := uc.Do(context.Background(), Command{ID: "", Version: 1})
	if !errors.Is(err, ErrInvalidID) {
		t.Fatalf("err = %v, want ErrInvalidID", err)
	}
}

func TestDo_normalizeError_zeroVersion(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{}
	clk := mockClocker{now: time.Now()}
	uc := New(runner, clk)

	err := uc.Do(context.Background(), Command{ID: "some-id", Version: 0})
	if !errors.Is(err, ErrInvalidVersion) {
		t.Fatalf("err = %v, want ErrInvalidVersion", err)
	}
}

func TestDo_findByIDError(t *testing.T) {
	t.Parallel()

	taskID := dtask.NewTaskID()
	findErr := errors.New("db connection lost")

	taskRepo := &mockTaskRepo{findErr: findErr}
	eventRepo := &mockEventRepo{}
	deps := &mockDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockRunner{deps: deps}
	clk := mockClocker{now: time.Now()}

	uc := New(runner, clk)
	err := uc.Do(context.Background(), Command{UserID: testUID, ID: taskID.Value(), Version: 1})
	if !errors.Is(err, findErr) {
		t.Fatalf("err = %v, want %v", err, findErr)
	}
	if eventRepo.insertCalled {
		t.Fatal("Insert should not be called on FindByID error")
	}
}

func TestDo_deleteError(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	taskID := dtask.NewTaskID()
	delErr := errors.New("delete failed")

	taskRepo := &mockTaskRepo{
		findTask:  testTask(taskID, now.Add(7*24*time.Hour)),
		deleteErr: delErr,
	}
	eventRepo := &mockEventRepo{}
	deps := &mockDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockRunner{deps: deps}
	clk := mockClocker{now: now}

	uc := New(runner, clk)
	err := uc.Do(context.Background(), Command{UserID: testUID, ID: taskID.Value(), Version: 1})
	if !errors.Is(err, delErr) {
		t.Fatalf("err = %v, want %v", err, delErr)
	}
	if !eventRepo.insertCalled {
		t.Fatal("Insert should be called before Delete")
	}
}

func TestDo_insertEventError(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	taskID := dtask.NewTaskID()
	insertErr := errors.New("event insert failed")

	taskRepo := &mockTaskRepo{findTask: testTask(taskID, now.Add(7*24*time.Hour))}
	eventRepo := &mockEventRepo{insertErr: insertErr}
	deps := &mockDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockRunner{deps: deps}
	clk := mockClocker{now: now}

	uc := New(runner, clk)
	err := uc.Do(context.Background(), Command{UserID: testUID, ID: taskID.Value(), Version: 1})
	if !errors.Is(err, insertErr) {
		t.Fatalf("err = %v, want %v", err, insertErr)
	}
}
