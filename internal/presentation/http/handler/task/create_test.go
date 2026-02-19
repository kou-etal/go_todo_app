package task

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kou-etal/go_todo_app/internal/usecase/task/create"
)

func newCreateHandler(taskRepo *mockTaskRepo, eventRepo *mockEventRepo) http.Handler {
	deps := &mockTaskEventDeps{taskRepo: taskRepo, taskEventRepo: eventRepo}
	runner := &mockTaskEventRunner{deps: deps}
	clk := stubClocker{now: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)}
	uc := create.New(runner, clk)
	return NewCreate(uc, stubLogger{})
}

func TestCreateHandler_happyPath(t *testing.T) {
	t.Parallel()

	h := newCreateHandler(&mockTaskRepo{}, &mockEventRepo{})
	body := `{"title":"test task","description":"desc","due_date":7}`
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["id"] == "" {
		t.Fatal("response id should not be empty")
	}
}

func TestCreateHandler_invalidJSON(t *testing.T) {
	t.Parallel()

	h := newCreateHandler(&mockTaskRepo{}, &mockEventRepo{})
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["message"] != "invalid body" {
		t.Fatalf("message = %q, want %q", resp["message"], "invalid body")
	}
}

func TestCreateHandler_emptyTitle(t *testing.T) {
	t.Parallel()

	h := newCreateHandler(&mockTaskRepo{}, &mockEventRepo{})
	body := `{"title":"","description":"desc","due_date":7}`
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["message"] != "title is required" {
		t.Fatalf("message = %q, want %q", resp["message"], "title is required")
	}
}

func TestCreateHandler_invalidDue(t *testing.T) {
	t.Parallel()

	h := newCreateHandler(&mockTaskRepo{}, &mockEventRepo{})
	body := `{"title":"task","description":"desc","due_date":5}`
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["message"] != "invalid due_date" {
		t.Fatalf("message = %q, want %q", resp["message"], "invalid due_date")
	}
}

func TestCreateHandler_internalError(t *testing.T) {
	t.Parallel()

	runner := &mockTaskEventRunner{withErr: errInternal}
	clk := stubClocker{now: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)}
	uc := create.New(runner, clk)
	h := NewCreate(uc, stubLogger{})

	body := `{"title":"task","description":"desc","due_date":7}`
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["message"] != "internal server error" {
		t.Fatalf("message = %q, want %q", resp["message"], "internal server error")
	}
}
