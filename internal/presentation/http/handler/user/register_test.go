package userhandler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kou-etal/go_todo_app/internal/logger"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
	"github.com/kou-etal/go_todo_app/internal/usecase/user/register"

	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

// --- stub logger ---

type stubLogger struct{}

func (stubLogger) Debug(_ context.Context, _ string, _ error, _ ...logger.Attr) {}
func (stubLogger) Info(_ context.Context, _ string, _ ...logger.Attr)            {}
func (stubLogger) Warn(_ context.Context, _ string, _ ...logger.Attr)            {}
func (stubLogger) Error(_ context.Context, _ string, _ error, _ ...logger.Attr)  {}

// --- stub clocker ---

type stubClocker struct{ now time.Time }

func (c stubClocker) Now() time.Time { return c.now }

// --- mock runner ---

type mockRunner struct {
	deps    usetx.RegisterDeps
	withErr error
}

func (m *mockRunner) WithinTx(ctx context.Context, fn func(ctx context.Context, deps usetx.RegisterDeps) error) error {
	if m.withErr != nil {
		return m.withErr
	}
	return fn(ctx, m.deps)
}

// --- mock register deps ---

type mockRegisterDeps struct {
	userRepo        duser.UserRepository
	emailVerifyRepo dverify.EmailVerificationRepository
}

func (m *mockRegisterDeps) UserRepo() duser.UserRepository                     { return m.userRepo }
func (m *mockRegisterDeps) EmailVerifyRepo() dverify.EmailVerificationRepository { return m.emailVerifyRepo }

// --- mock user repo ---

type mockUserRepo struct {
	storeErr error
}

func (m *mockUserRepo) Store(_ context.Context, _ *duser.User) error { return m.storeErr }

// --- mock email verify repo ---

type mockEmailVerifyRepo struct {
	insertErr error
}

func (m *mockEmailVerifyRepo) Insert(_ context.Context, _ *dverify.EmailVerificationToken) error {
	return m.insertErr
}

// --- stubs ---

type stubHasher struct {
	hash string
	err  error
}

func (s stubHasher) Hash(_ string) (string, error) { return s.hash, s.err }
func (s stubHasher) Compare(_, _ string) error      { return nil }

type stubTokenGenerator struct {
	token string
	err   error
}

func (s stubTokenGenerator) Generate() (string, error) { return s.token, s.err }

type stubTokenHasher struct {
	hash string
	err  error
}

func (s stubTokenHasher) Hash(_ string) (string, error) { return s.hash, s.err }

// --- helpers ---

func newRegisterHandler(userRepo *mockUserRepo, emailRepo *mockEmailVerifyRepo) http.Handler {
	deps := &mockRegisterDeps{userRepo: userRepo, emailVerifyRepo: emailRepo}
	runner := &mockRunner{deps: deps}
	clk := stubClocker{now: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)}
	uc := register.New(
		runner, clk,
		stubHasher{hash: "$2a$10$hashed"},
		stubTokenGenerator{token: "plain-token"},
		stubTokenHasher{hash: "sha256-hash"},
	)
	return NewRegister(uc, stubLogger{})
}

// --- tests ---

func TestRegisterHandler_happyPath(t *testing.T) {
	t.Parallel()

	h := newRegisterHandler(&mockUserRepo{}, &mockEmailVerifyRepo{})
	body := `{"email":"test@example.com","password":"password1234","user_name":"alice"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
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

func TestRegisterHandler_invalidJSON(t *testing.T) {
	t.Parallel()

	h := newRegisterHandler(&mockUserRepo{}, &mockEmailVerifyRepo{})
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader("{invalid"))
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

func TestRegisterHandler_emptyEmail(t *testing.T) {
	t.Parallel()

	h := newRegisterHandler(&mockUserRepo{}, &mockEmailVerifyRepo{})
	body := `{"email":"","password":"password1234","user_name":"alice"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["message"] != "email is required" {
		t.Fatalf("message = %q, want %q", resp["message"], "email is required")
	}
}

func TestRegisterHandler_conflict(t *testing.T) {
	t.Parallel()

	userRepo := &mockUserRepo{storeErr: duser.ErrConflict}
	h := newRegisterHandler(userRepo, &mockEmailVerifyRepo{})

	body := `{"email":"test@example.com","password":"password1234","user_name":"alice"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["message"] != "conflict" {
		t.Fatalf("message = %q, want %q", resp["message"], "conflict")
	}
}

func TestRegisterHandler_internalError(t *testing.T) {
	t.Parallel()

	runner := &mockRunner{withErr: errors.New("internal")}
	clk := stubClocker{now: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)}
	uc := register.New(
		runner, clk,
		stubHasher{hash: "$2a$10$hashed"},
		stubTokenGenerator{token: "plain-token"},
		stubTokenHasher{hash: "sha256-hash"},
	)
	h := NewRegister(uc, stubLogger{})

	body := `{"email":"test@example.com","password":"password1234","user_name":"alice"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
