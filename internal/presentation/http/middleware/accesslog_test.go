package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseRecorder_capturesStatus(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	rec := &responseRecorder{w: w, statusCode: http.StatusOK}

	rec.WriteHeader(http.StatusCreated)

	if rec.statusCode != http.StatusCreated {
		t.Fatalf("statusCode = %d, want %d", rec.statusCode, http.StatusCreated)
	}

	body := []byte("hello")
	n, err := rec.Write(body)
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != len(body) {
		t.Fatalf("written = %d, want %d", n, len(body))
	}
	if rec.bytes != int64(len(body)) {
		t.Fatalf("bytes = %d, want %d", rec.bytes, len(body))
	}
}

func TestResponseRecorder_defaultStatus200(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	rec := &responseRecorder{w: w, statusCode: http.StatusOK}

	body := []byte("hello")
	_, err := rec.Write(body)
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}

	if rec.statusCode != http.StatusOK {
		t.Fatalf("statusCode = %d, want %d", rec.statusCode, http.StatusOK)
	}
	if !rec.wroteHeader {
		t.Fatal("wroteHeader should be true after Write")
	}
}

func TestResponseRecorder_doubleWriteHeaderIgnored(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	rec := &responseRecorder{w: w, statusCode: http.StatusOK}

	rec.WriteHeader(http.StatusCreated)
	rec.WriteHeader(http.StatusNotFound)

	if rec.statusCode != http.StatusCreated {
		t.Fatalf("statusCode = %d, want %d (second WriteHeader should be ignored)", rec.statusCode, http.StatusCreated)
	}
}
