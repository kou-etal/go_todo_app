package task

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	"github.com/kou-etal/go_todo_app/internal/domain/user"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/update"
)

func newUpdateHandler(taskRepo *mockTaskRepo, eventRepo *mockEventRepo) http.Handler {
	deps := &mockTaskEventDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockTaskEventRunner{deps: deps}
	clk := stubClocker{now: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)}
	uc := update.New(runner, clk)
	return NewUpdate(uc, stubLogger{})
}

func buildUpdateRequest(t *testing.T, id, body string) *http.Request {
	t.Helper()
	req := withAuthContext(httptest.NewRequest(http.MethodPut, "/tasks/"+id, strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", id)
	return req
}

func existingTask() *dtask.Task {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	title, _ := dtask.NewTaskTitle("old title")
	desc, _ := dtask.NewTaskDescription("old desc")
	due, _ := dtask.NewDueDateFromOption(now, dtask.Due7Days)
	return dtask.NewTask(user.UserID(testUserID), title, desc, due, now)
}

func TestUpdateHandler_happyPath(t *testing.T) {
	t.Parallel()

	task := existingTask()
	taskRepo := &mockTaskRepo{findTask: task}
	h := newUpdateHandler(taskRepo, &mockEventRepo{})

	body := `{"version":1,"title":"new title"}`
	req := buildUpdateRequest(t, task.ID().Value(), body)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["id"] == "" {
		t.Fatal("response id should not be empty")
	}
}

func TestUpdateHandler_zeroVersion(t *testing.T) {
	t.Parallel()

	h := newUpdateHandler(&mockTaskRepo{}, &mockEventRepo{})
	body := `{"version":0,"title":"new title"}`
	req := buildUpdateRequest(t, "550e8400-e29b-41d4-a716-446655440000", body)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["message"] != "invalid version" {
		t.Fatalf("message = %q, want %q", resp["message"], "invalid version")
	}
}

func TestUpdateHandler_conflict(t *testing.T) {
	t.Parallel()

	task := existingTask()
	taskRepo := &mockTaskRepo{findTask: task}
	h := newUpdateHandler(taskRepo, &mockEventRepo{})

	// version 999 != task.Version() (1) → conflict
	body := `{"version":999,"title":"new title"}`
	req := buildUpdateRequest(t, task.ID().Value(), body)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["message"] != "conflict" {
		t.Fatalf("message = %q, want %q", resp["message"], "conflict")
	}
}

func TestUpdateHandler_internalError(t *testing.T) {
	t.Parallel()

	runner := &mockTaskEventRunner{withErr: errInternal}
	clk := stubClocker{now: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)}
	uc := update.New(runner, clk)
	h := NewUpdate(uc, stubLogger{})

	body := `{"version":1,"title":"new title"}`
	req := buildUpdateRequest(t, "550e8400-e29b-41d4-a716-446655440000", body)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
