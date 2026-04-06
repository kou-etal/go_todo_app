package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kou-etal/go_todo_app/internal/auth"
	"github.com/kou-etal/go_todo_app/internal/logger"
)

// --- stub logger ---

type stubLogger struct{}

func (stubLogger) Debug(_ context.Context, _ string, _ error, _ ...logger.Attr) {}
func (stubLogger) Info(_ context.Context, _ string, _ ...logger.Attr)            {}
func (stubLogger) Warn(_ context.Context, _ string, _ ...logger.Attr)            {}
func (stubLogger) Error(_ context.Context, _ string, _ error, _ ...logger.Attr)  {}

// --- stub parser ---

type stubParser struct {
	userID string
	err    error
}

func (s stubParser) ParseAccessToken(_ string) (string, error) {
	return s.userID, s.err
}

// --- tests ---

func TestAuth_noHeader_returns401(t *testing.T) {
	t.Parallel()

	parser := stubParser{userID: "00000000-0000-0000-0000-000000000001"}
	mw := Auth(parser, stubLogger{})

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()

	mw(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["message"] != "authorization header is required" {
		t.Fatalf("message = %q", resp["message"])
	}
}

func TestAuth_invalidFormat_returns401(t *testing.T) {
	t.Parallel()

	parser := stubParser{userID: "00000000-0000-0000-0000-000000000001"}
	mw := Auth(parser, stubLogger{})

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Authorization", "Basic abc123")
	rec := httptest.NewRecorder()

	mw(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["message"] != "invalid authorization header format" {
		t.Fatalf("message = %q", resp["message"])
	}
}

func TestAuth_invalidToken_returns401(t *testing.T) {
	t.Parallel()

	parser := stubParser{err: errors.New("token expired")}
	mw := Auth(parser, stubLogger{})

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	mw(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["message"] != "invalid or expired token" {
		t.Fatalf("message = %q", resp["message"])
	}
}

func TestAuth_validToken_passesThrough(t *testing.T) {
	t.Parallel()

	uid := "00000000-0000-0000-0000-000000000001"
	parser := stubParser{userID: uid}
	mw := Auth(parser, stubLogger{})

	var gotUserID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			t.Fatal("user id not found in context")
		}
		gotUserID = id.Value()
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	mw(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotUserID != uid {
		t.Fatalf("context userID = %q, want %q", gotUserID, uid)
	}
}
