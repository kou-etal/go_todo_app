package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kou-etal/go_todo_app/internal/observability/requestid"
)

func TestRequestID_generatesNewID(t *testing.T) {
	t.Parallel()

	var ctxID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := requestid.FromContext(r.Context())
		if !ok {
			t.Fatal("request id not found in context")
		}
		ctxID = id
		w.WriteHeader(http.StatusOK)
	})

	h := RequestID(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	headerID := rec.Header().Get(HeaderRequestID)
	if headerID == "" {
		t.Fatal("X-Request-Id response header should not be empty")
	}
	if ctxID != headerID {
		t.Fatalf("context id = %q, header id = %q, should match", ctxID, headerID)
	}
}

func TestRequestID_respectsExisting(t *testing.T) {
	t.Parallel()

	existing := "my-custom-request-id"

	var ctxID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := requestid.FromContext(r.Context())
		ctxID = id
		w.WriteHeader(http.StatusOK)
	})

	h := RequestID(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderRequestID, existing)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Header().Get(HeaderRequestID) != existing {
		t.Fatalf("header = %q, want %q", rec.Header().Get(HeaderRequestID), existing)
	}
	if ctxID != existing {
		t.Fatalf("context id = %q, want %q", ctxID, existing)
	}
}
