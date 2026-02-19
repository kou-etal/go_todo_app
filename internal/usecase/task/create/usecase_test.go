package create

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	taskevent "github.com/kou-etal/go_todo_app/internal/domain/event"
	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
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
	storeErr    error
	storeCalled bool
}

func (m *mockTaskRepo) Store(_ context.Context, _ *dtask.Task) error {
	m.storeCalled = true
	return m.storeErr
}
func (m *mockTaskRepo) FindByID(_ context.Context, _ dtask.TaskID) (*dtask.Task, error) {
	return nil, nil
}
func (m *mockTaskRepo) Update(_ context.Context, _ *dtask.Task) error         { return nil }
func (m *mockTaskRepo) Delete(_ context.Context, _ dtask.TaskID, _ uint64) error { return nil }
func (m *mockTaskRepo) List(_ context.Context, _ dtask.ListQuery) ([]*dtask.Task, *dtask.ListCursor, error) {
	return nil, nil, nil
}

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

// --- tests ---

func TestDo_happyPath(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	taskRepo := &mockTaskRepo{}
	eventRepo := &mockEventRepo{}
	deps := &mockDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockRunner{deps: deps}
	clk := mockClocker{now: now}

	uc := New(runner, clk)
	res, err := uc.Do(context.Background(), Command{
		Title:       "test task",
		Description: "test desc",
		DueDate:     7,
	})
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	if res.ID == "" {
		t.Fatal("result.ID should not be empty")
	}
	if !taskRepo.storeCalled {
		t.Fatal("Store was not called")
	}
	if !eventRepo.insertCalled {
		t.Fatal("Insert was not called")
	}
	if eventRepo.insertedEvent.EventType() != taskevent.EventCreated {
		t.Fatalf("event type = %s, want created", eventRepo.insertedEvent.EventType())
	}
}

func TestDo_emptyTitle(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{}
	clk := mockClocker{now: time.Now()}
	uc := New(runner, clk)

	_, err := uc.Do(context.Background(), Command{
		Title:       "",
		Description: "desc",
		DueDate:     7,
	})
	if !errors.Is(err, ErrEmptyTitle) {
		t.Fatalf("err = %v, want ErrEmptyTitle", err)
	}
}

func TestDo_titleByteLimit(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{}
	clk := mockClocker{now: time.Now()}
	uc := New(runner, clk)

	_, err := uc.Do(context.Background(), Command{
		Title:       strings.Repeat("a", 1025),
		Description: "desc",
		DueDate:     7,
	})
	if !errors.Is(err, ErrTitleTooLong) {
		t.Fatalf("err = %v, want ErrTitleTooLong", err)
	}
}

func TestDo_descByteLimit(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{}
	clk := mockClocker{now: time.Now()}
	uc := New(runner, clk)

	_, err := uc.Do(context.Background(), Command{
		Title:       "title",
		Description: strings.Repeat("a", 16385),
		DueDate:     7,
	})
	if !errors.Is(err, ErrDescriptionTooLong) {
		t.Fatalf("err = %v, want ErrDescriptionTooLong", err)
	}
}

func TestDo_invalidDue(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{}
	clk := mockClocker{now: time.Now()}
	uc := New(runner, clk)

	_, err := uc.Do(context.Background(), Command{
		Title:       "title",
		Description: "desc",
		DueDate:     5,
	})
	if !errors.Is(err, ErrInvalidDueOption) {
		t.Fatalf("err = %v, want ErrInvalidDueOption", err)
	}
}

func TestDo_storeError(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	storeErr := errors.New("store failed")
	taskRepo := &mockTaskRepo{storeErr: storeErr}
	eventRepo := &mockEventRepo{}
	deps := &mockDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockRunner{deps: deps}
	clk := mockClocker{now: now}

	uc := New(runner, clk)
	_, err := uc.Do(context.Background(), Command{
		Title:       "test task",
		Description: "test desc",
		DueDate:     7,
	})
	if !errors.Is(err, storeErr) {
		t.Fatalf("err = %v, want %v", err, storeErr)
	}
	if eventRepo.insertCalled {
		t.Fatal("Insert should not be called on Store error")
	}
}

func TestDo_insertEventError(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	insertErr := errors.New("event insert failed")
	taskRepo := &mockTaskRepo{}
	eventRepo := &mockEventRepo{insertErr: insertErr}
	deps := &mockDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockRunner{deps: deps}
	clk := mockClocker{now: now}

	uc := New(runner, clk)
	_, err := uc.Do(context.Background(), Command{
		Title:       "test task",
		Description: "test desc",
		DueDate:     7,
	})
	if !errors.Is(err, insertErr) {
		t.Fatalf("err = %v, want %v", err, insertErr)
	}
}
