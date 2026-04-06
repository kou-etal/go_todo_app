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

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	drefresh "github.com/kou-etal/go_todo_app/internal/domain/user/refresh"
	"github.com/kou-etal/go_todo_app/internal/usecase/user/login"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

// --- mock login runner ---

type mockLoginRunner struct {
	deps    usetx.LoginDeps
	withErr error
}

func (m *mockLoginRunner) WithinTx(ctx context.Context, fn func(ctx context.Context, deps usetx.LoginDeps) error) error {
	if m.withErr != nil {
		return m.withErr
	}
	return fn(ctx, m.deps)
}

// --- mock login deps ---

type mockLoginDeps struct {
	userRepo         duser.UserRepository
	refreshTokenRepo drefresh.RefreshTokenRepository
}

func (m *mockLoginDeps) UserRepo() duser.UserRepository                { return m.userRepo }
func (m *mockLoginDeps) RefreshTokenRepo() drefresh.RefreshTokenRepository { return m.refreshTokenRepo }

// --- mock login user repo ---

type mockLoginUserRepo struct {
	findUser *duser.User
	findErr  error
}

func (m *mockLoginUserRepo) Store(_ context.Context, _ *duser.User) error { return nil }
func (m *mockLoginUserRepo) FindByEmail(_ context.Context, _ duser.UserEmail) (*duser.User, error) {
	return m.findUser, m.findErr
}

// --- mock refresh token repo ---

type mockLoginRefreshRepo struct{}

func (m *mockLoginRefreshRepo) Store(_ context.Context, _ *drefresh.RefreshToken) error { return nil }
func (m *mockLoginRefreshRepo) FindByTokenHashForUpdate(_ context.Context, _ drefresh.TokenHash) (*drefresh.RefreshToken, error) {
	return nil, nil
}
func (m *mockLoginRefreshRepo) Update(_ context.Context, _ *drefresh.RefreshToken) error { return nil }

// --- stubs ---

type stubAccessTokenGen struct {
	token string
	err   error
}

func (s stubAccessTokenGen) GenerateAccessToken(_ string) (string, error) { return s.token, s.err }

type stubRefreshHasher struct {
	hash string
	err  error
}

func (s stubRefreshHasher) Hash(_ string) (string, error) { return s.hash, s.err }

type stubRefreshGen struct {
	token string
	err   error
}

func (s stubRefreshGen) Generate() (string, error) { return s.token, s.err }

type stubPWHasher struct {
	hash       string
	hashErr    error
	compareErr error
}

func (s stubPWHasher) Hash(_ string) (string, error) { return s.hash, s.hashErr }
func (s stubPWHasher) Compare(_, _ string) error      { return s.compareErr }

// --- helpers ---

const loginTestUID = "00000000-0000-0000-0000-000000000001"

func loginTestUser() *duser.User {
	email, _ := duser.NewUserEmail("test@example.com")
	pw, _ := duser.ReconstructUserPassword("$2a$10$hashed")
	name, _ := duser.NewUserName("alice")
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return duser.ReconstructUser(duser.UserID(loginTestUID), email, pw, name, nil, now, now, 1)
}

func newLoginHandler(userRepo *mockLoginUserRepo) http.Handler {
	refreshRepo := &mockLoginRefreshRepo{}
	deps := &mockLoginDeps{userRepo: userRepo, refreshTokenRepo: refreshRepo}
	runner := &mockLoginRunner{deps: deps}
	clk := stubClocker{now: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)}
	uc := login.New(
		runner, clk,
		stubPWHasher{hash: "$2a$10$hashed"},
		stubAccessTokenGen{token: "access-token-123"},
		stubRefreshHasher{hash: "sha256-hash"},
		stubRefreshGen{token: "refresh-token-456"},
		7*24*time.Hour,
		900,
	)
	return NewLogin(uc, stubLogger{})
}

// --- tests ---

func TestLoginHandler_happyPath(t *testing.T) {
	t.Parallel()

	h := newLoginHandler(&mockLoginUserRepo{findUser: loginTestUser()})
	body := `{"email":"test@example.com","password":"password1234"}`
	req := httptest.NewRequest(http.MethodPost, "/users/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["access_token"] == "" {
		t.Fatal("access_token should not be empty")
	}
	if resp["refresh_token"] == "" {
		t.Fatal("refresh_token should not be empty")
	}
	if resp["expires_in"].(float64) != 900 {
		t.Fatalf("expires_in = %v, want 900", resp["expires_in"])
	}
}

func TestLoginHandler_invalidCredentials(t *testing.T) {
	t.Parallel()

	h := newLoginHandler(&mockLoginUserRepo{findErr: duser.ErrNotFound})
	body := `{"email":"unknown@example.com","password":"password1234"}`
	req := httptest.NewRequest(http.MethodPost, "/users/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["message"] != "invalid credentials" {
		t.Fatalf("message = %q, want %q", resp["message"], "invalid credentials")
	}
}

func TestLoginHandler_wrongPassword(t *testing.T) {
	t.Parallel()

	userRepo := &mockLoginUserRepo{findUser: loginTestUser()}
	refreshRepo := &mockLoginRefreshRepo{}
	deps := &mockLoginDeps{userRepo: userRepo, refreshTokenRepo: refreshRepo}
	runner := &mockLoginRunner{deps: deps}
	clk := stubClocker{now: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)}
	uc := login.New(
		runner, clk,
		stubPWHasher{compareErr: errors.New("mismatch")},
		stubAccessTokenGen{token: "access-token-123"},
		stubRefreshHasher{hash: "sha256-hash"},
		stubRefreshGen{token: "refresh-token-456"},
		7*24*time.Hour,
		900,
	)
	h := NewLogin(uc, stubLogger{})

	body := `{"email":"test@example.com","password":"wrong-password"}`
	req := httptest.NewRequest(http.MethodPost, "/users/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestLoginHandler_emptyEmail(t *testing.T) {
	t.Parallel()

	h := newLoginHandler(&mockLoginUserRepo{})
	body := `{"email":"","password":"password1234"}`
	req := httptest.NewRequest(http.MethodPost, "/users/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["message"] != "email is required" {
		t.Fatalf("message = %q, want %q", resp["message"], "email is required")
	}
}

func TestLoginHandler_invalidJSON(t *testing.T) {
	t.Parallel()

	h := newLoginHandler(&mockLoginUserRepo{})
	req := httptest.NewRequest(http.MethodPost, "/users/login", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
