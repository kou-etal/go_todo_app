package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kou-etal/go_todo_app/internal/auth"
	"github.com/kou-etal/go_todo_app/internal/domain/user"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
)

// --- stub auth middleware (JWT検証を模倣) ---

func stubAuthMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header != "Bearer valid-token" {
			responder.JSON(w, http.StatusUnauthorized, responder.ErrResponse{Message: "unauthorized"})
			return
		}
		ctx := auth.WithUserID(r.Context(), user.UserID("00000000-0000-0000-0000-000000000001"))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// --- stub handler ---

func stubHandler(status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	})
}

func testRouter() http.Handler {
	return New(Deps{
		Task: TaskDeps{
			List:   stubHandler(http.StatusOK),
			Create: stubHandler(http.StatusCreated),
			Update: stubHandler(http.StatusOK),
			Delete: stubHandler(http.StatusNoContent),
		},
		User: UserDeps{
			Register: stubHandler(http.StatusCreated),
			Login:    stubHandler(http.StatusOK),
			Refresh:  stubHandler(http.StatusOK),
		},
		AuthMW: stubAuthMW,
	})
}

// --- tests ---

func TestTasks_noAuth_returns401(t *testing.T) {
	t.Parallel()

	h := testRouter()

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /tasks without auth: status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["message"] != "unauthorized" {
		t.Fatalf("message = %q, want %q", resp["message"], "unauthorized")
	}
}

func TestTasksSlash_noAuth_returns401(t *testing.T) {
	t.Parallel()

	h := testRouter()

	req := httptest.NewRequest(http.MethodPut, "/tasks/some-id", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("PUT /tasks/some-id without auth: status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestTasks_withAuth_passesThrough(t *testing.T) {
	t.Parallel()

	h := testRouter()

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /tasks with auth: status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestUsersLogin_noAuth_allowed(t *testing.T) {
	t.Parallel()

	h := testRouter()

	req := httptest.NewRequest(http.MethodPost, "/users/login", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /users/login should not require auth: status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestUsersRefresh_noAuth_allowed(t *testing.T) {
	t.Parallel()

	h := testRouter()

	req := httptest.NewRequest(http.MethodPost, "/users/refresh", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /users/refresh should not require auth: status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestUsersRegister_noAuth_allowed(t *testing.T) {
	t.Parallel()

	h := testRouter()

	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /users should not require auth: status = %d, want %d", rec.Code, http.StatusCreated)
	}
}
