package responder

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON_statusAndContentType(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	body := map[string]string{"key": "value"}

	JSON(rec, http.StatusCreated, body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want %q", ct, "application/json; charset=utf-8")
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["key"] != "value" {
		t.Fatalf("key = %q, want %q", resp["key"], "value")
	}
}

func TestJSON_errResponse(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	JSON(rec, http.StatusBadRequest, ErrResponse{Message: "bad request"})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["message"] != "bad request" {
		t.Fatalf("message = %q, want %q", resp["message"], "bad request")
	}
}
