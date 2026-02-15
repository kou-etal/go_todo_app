package update

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
	withErr error
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

	updateErr    error
	updateCalled bool

	deleteErr error
}

func (m *mockTaskRepo) FindByID(_ context.Context, id dtask.TaskID) (*dtask.Task, error) {
	return m.findTask, m.findErr
}

func (m *mockTaskRepo) Update(_ context.Context, _ *dtask.Task) error {
	m.updateCalled = true
	return m.updateErr
}

func (m *mockTaskRepo) Delete(_ context.Context, _ dtask.TaskID, _ uint64) error {
	return m.deleteErr
}
func (m *mockTaskRepo) List(_ context.Context, _ dtask.ListQuery) ([]*dtask.Task, *dtask.ListCursor, error) {
	return nil, nil, nil
}
func (m *mockTaskRepo) Store(_ context.Context, _ *dtask.Task) error { return nil }

// --- mock event repo ---

type mockEventRepo struct {
	insertErr     error
	insertCalled  bool
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

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func testTask(id dtask.TaskID, version uint64) *dtask.Task {
	title, _ := dtask.NewTaskTitle("old title")
	desc, _ := dtask.NewTaskDescription("old desc")
	due, _ := dtask.NewDueDateFromTime(time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC))
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return dtask.ReconstructTask(id, user.UserID("u-1"), title, desc, dtask.StatusTodo, due, now, now, version)
}

// --- tests ---

func TestDo_happyPath_titleOnly(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	taskID := dtask.NewTaskID()

	taskRepo := &mockTaskRepo{findTask: testTask(taskID, 1)}
	eventRepo := &mockEventRepo{}
	deps := &mockDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockRunner{deps: deps}
	clk := mockClocker{now: now}

	uc := New(runner, clk)
	result, err := uc.Do(context.Background(), Command{
		ID:      taskID.Value(),
		Version: 1,
		Title:   strPtr("new title"),
	})
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	if result.ID != taskID.Value() {
		t.Fatalf("result.ID = %s, want %s", result.ID, taskID.Value())
	}
	if !taskRepo.updateCalled {
		t.Fatal("Update was not called")
	}
	if !eventRepo.insertCalled {
		t.Fatal("Insert was not called")
	}
	if eventRepo.insertedEvent.EventType() != taskevent.EventUpdated {
		t.Fatalf("event type = %s, want updated", eventRepo.insertedEvent.EventType())
	}
}

func TestDo_happyPath_allFields(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	taskID := dtask.NewTaskID()

	taskRepo := &mockTaskRepo{findTask: testTask(taskID, 1)}
	eventRepo := &mockEventRepo{}
	deps := &mockDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockRunner{deps: deps}
	clk := mockClocker{now: now}

	uc := New(runner, clk)
	_, err := uc.Do(context.Background(), Command{
		ID:          taskID.Value(),
		Version:     1,
		Title:       strPtr("new title"),
		Description: strPtr("new desc"),
		DueDate:     intPtr(14),
	})
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	if !eventRepo.insertCalled {
		t.Fatal("Insert was not called")
	}
}

func TestDo_normalizeError_noFields(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{}
	clk := mockClocker{now: time.Now()}
	uc := New(runner, clk)

	_, err := uc.Do(context.Background(), Command{
		ID:      "some-id",
		Version: 1,
	})
	if !errors.Is(err, ErrNoFieldsToUpdate) {
		t.Fatalf("err = %v, want ErrNoFieldsToUpdate", err)
	}
}

func TestDo_versionConflict(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	taskID := dtask.NewTaskID()

	taskRepo := &mockTaskRepo{findTask: testTask(taskID, 1)} // version=1
	eventRepo := &mockEventRepo{}
	deps := &mockDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockRunner{deps: deps}
	clk := mockClocker{now: now}

	uc := New(runner, clk)
	_, err := uc.Do(context.Background(), Command{
		ID:      taskID.Value(),
		Version: 2, // mismatch
		Title:   strPtr("new title"),
	})
	if !errors.Is(err, dtask.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	if eventRepo.insertCalled {
		t.Fatal("Insert should not be called on version conflict")
	}
}

func TestDo_findByIDError(t *testing.T) {
	t.Parallel()

	taskID := dtask.NewTaskID()
	findErr := errors.New("db error")

	taskRepo := &mockTaskRepo{findErr: findErr}
	eventRepo := &mockEventRepo{}
	deps := &mockDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockRunner{deps: deps}
	clk := mockClocker{now: time.Now()}

	uc := New(runner, clk)
	_, err := uc.Do(context.Background(), Command{
		ID:      taskID.Value(),
		Version: 1,
		Title:   strPtr("new title"),
	})
	if !errors.Is(err, findErr) {
		t.Fatalf("err = %v, want %v", err, findErr)
	}
}

func TestDo_updateError(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	taskID := dtask.NewTaskID()
	updateErr := errors.New("update failed")

	taskRepo := &mockTaskRepo{
		findTask:  testTask(taskID, 1),
		updateErr: updateErr,
	}
	eventRepo := &mockEventRepo{}
	deps := &mockDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockRunner{deps: deps}
	clk := mockClocker{now: now}

	uc := New(runner, clk)
	_, err := uc.Do(context.Background(), Command{
		ID:      taskID.Value(),
		Version: 1,
		Title:   strPtr("new title"),
	})
	if !errors.Is(err, updateErr) {
		t.Fatalf("err = %v, want %v", err, updateErr)
	}
	if eventRepo.insertCalled {
		t.Fatal("Insert should not be called on Update error")
	}
}

func TestDo_insertEventError(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	taskID := dtask.NewTaskID()
	insertErr := errors.New("event insert failed")

	taskRepo := &mockTaskRepo{findTask: testTask(taskID, 1)}
	eventRepo := &mockEventRepo{insertErr: insertErr}
	deps := &mockDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockRunner{deps: deps}
	clk := mockClocker{now: now}

	uc := New(runner, clk)
	_, err := uc.Do(context.Background(), Command{
		ID:      taskID.Value(),
		Version: 1,
		Title:   strPtr("new title"),
	})
	if !errors.Is(err, insertErr) {
		t.Fatalf("err = %v, want %v", err, insertErr)
	}
}
