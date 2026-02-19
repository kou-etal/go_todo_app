package task

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	remove "github.com/kou-etal/go_todo_app/internal/usecase/task/delete"
)

func newDeleteHandler(taskRepo *mockTaskRepo, eventRepo *mockEventRepo) http.Handler {
	deps := &mockTaskEventDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockTaskEventRunner{deps: deps}
	clk := stubClocker{now: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)}
	uc := remove.New(runner, clk)
	return NewDelete(uc, stubLogger{})
}

func buildDeleteRequest(t *testing.T, id, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodDelete, "/tasks/"+id, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", id)
	return req
}

func TestDeleteHandler_happyPath(t *testing.T) {
	t.Parallel()

	task := existingTask()
	taskRepo := &mockTaskRepo{findTask: task}
	h := newDeleteHandler(taskRepo, &mockEventRepo{})

	body := `{"version":1}`
	req := buildDeleteRequest(t, task.ID().Value(), body)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestDeleteHandler_emptyPathID(t *testing.T) {
	t.Parallel()

	h := newDeleteHandler(&mockTaskRepo{}, &mockEventRepo{})
	body := `{"version":1}`
	req := httptest.NewRequest(http.MethodDelete, "/tasks/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// PathValue("id") returns "" because no path value set
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDeleteHandler_notFound(t *testing.T) {
	t.Parallel()

	taskRepo := &mockTaskRepo{findErr: dtask.ErrNotFound}
	h := newDeleteHandler(taskRepo, &mockEventRepo{})

	body := `{"version":1}`
	req := buildDeleteRequest(t, "550e8400-e29b-41d4-a716-446655440000", body)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestDeleteHandler_conflict(t *testing.T) {
	t.Parallel()

	task := existingTask()
	taskRepo := &mockTaskRepo{findTask: task, deleteErr: dtask.ErrConflict}
	h := newDeleteHandler(taskRepo, &mockEventRepo{})

	body := `{"version":1}`
	req := buildDeleteRequest(t, task.ID().Value(), body)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}
