package task

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kou-etal/go_todo_app/internal/usecase/task/list"
)

func newListHandler(taskRepo *mockTaskRepo) http.Handler {
	uc := list.New(taskRepo)
	return NewList(uc, stubLogger{})
}

func TestListHandler_happyPath(t *testing.T) {
	t.Parallel()

	taskRepo := &mockTaskRepo{listTasks: nil, listNext: nil}
	h := newListHandler(taskRepo)

	req := httptest.NewRequest(http.MethodGet, "/tasks?limit=10&sort=created", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var resp map[string]json.RawMessage
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if _, ok := resp["items"]; !ok {
		t.Fatal("response should contain items key")
	}
}

func TestListHandler_invalidLimitNaN(t *testing.T) {
	t.Parallel()

	h := newListHandler(&mockTaskRepo{})
	req := httptest.NewRequest(http.MethodGet, "/tasks?limit=abc", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["message"] != "invalid limit" {
		t.Fatalf("message = %q, want %q", resp["message"], "invalid limit")
	}
}

func TestListHandler_invalidSort(t *testing.T) {
	t.Parallel()

	h := newListHandler(&mockTaskRepo{})
	req := httptest.NewRequest(http.MethodGet, "/tasks?sort=bad_sort", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["message"] != "invalid sort" {
		t.Fatalf("message = %q, want %q", resp["message"], "invalid sort")
	}
}

func TestListHandler_internalError(t *testing.T) {
	t.Parallel()

	taskRepo := &mockTaskRepo{listErr: errors.New("db down")}
	h := newListHandler(taskRepo)

	req := httptest.NewRequest(http.MethodGet, "/tasks?limit=10&sort=created", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
