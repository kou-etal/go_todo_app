package userhandler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	drefresh "github.com/kou-etal/go_todo_app/internal/domain/user/refresh"
	"github.com/kou-etal/go_todo_app/internal/usecase/user/refresh"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

// --- mock refresh runner ---

type mockRefreshRunner struct {
	deps    usetx.RefreshDeps
	withErr error
}

func (m *mockRefreshRunner) WithinTx(ctx context.Context, fn func(ctx context.Context, deps usetx.RefreshDeps) error) error {
	if m.withErr != nil {
		return m.withErr
	}
	return fn(ctx, m.deps)
}

// --- mock refresh deps ---

type mockRefreshDeps struct {
	refreshTokenRepo drefresh.RefreshTokenRepository
}

func (m *mockRefreshDeps) RefreshTokenRepo() drefresh.RefreshTokenRepository { return m.refreshTokenRepo }

// --- mock refresh token repo ---

type mockRefreshTokenRepo struct {
	findToken *drefresh.RefreshToken
	findErr   error
}

func (m *mockRefreshTokenRepo) FindByTokenHashForUpdate(_ context.Context, _ drefresh.TokenHash) (*drefresh.RefreshToken, error) {
	return m.findToken, m.findErr
}
func (m *mockRefreshTokenRepo) Update(_ context.Context, _ *drefresh.RefreshToken) error { return nil }
func (m *mockRefreshTokenRepo) Store(_ context.Context, _ *drefresh.RefreshToken) error  { return nil }

// --- helpers ---

const refreshTestUID = "00000000-0000-0000-0000-000000000002"

func refreshValidToken(now time.Time) *drefresh.RefreshToken {
	tokenHash, _ := drefresh.ReconstructTokenHash("sha256-old-hash")
	return drefresh.ReconstructRefreshToken(
		drefresh.NewTokenID(),
		duser.UserID(refreshTestUID),
		tokenHash,
		now.Add(7*24*time.Hour),
		nil,
		now.Add(-1*time.Hour),
	)
}

func newRefreshHandler(repo *mockRefreshTokenRepo) http.Handler {
	deps := &mockRefreshDeps{refreshTokenRepo: repo}
	runner := &mockRefreshRunner{deps: deps}
	clk := stubClocker{now: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)}
	uc := refresh.New(
		runner, clk,
		stubAccessTokenGen{token: "new-access-token"},
		stubRefreshHasher{hash: "sha256-hash"},
		stubRefreshGen{token: "new-refresh-token"},
		7*24*time.Hour,
		900,
	)
	return NewRefresh(uc, stubLogger{})
}

// --- tests ---

func TestRefreshHandler_happyPath(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	repo := &mockRefreshTokenRepo{findToken: refreshValidToken(now)}
	h := newRefreshHandler(repo)

	body := `{"refresh_token":"old-plain-token"}`
	req := httptest.NewRequest(http.MethodPost, "/users/refresh", strings.NewReader(body))
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
}

func TestRefreshHandler_emptyToken(t *testing.T) {
	t.Parallel()

	h := newRefreshHandler(&mockRefreshTokenRepo{})
	body := `{"refresh_token":""}`
	req := httptest.NewRequest(http.MethodPost, "/users/refresh", strings.NewReader(body))
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
	if resp["message"] != "refresh token is required" {
		t.Fatalf("message = %q, want %q", resp["message"], "refresh token is required")
	}
}

func TestRefreshHandler_invalidToken(t *testing.T) {
	t.Parallel()

	repo := &mockRefreshTokenRepo{findErr: drefresh.ErrNotFound}
	h := newRefreshHandler(repo)

	body := `{"refresh_token":"unknown-token"}`
	req := httptest.NewRequest(http.MethodPost, "/users/refresh", strings.NewReader(body))
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
	if resp["message"] != "invalid or expired refresh token" {
		t.Fatalf("message = %q, want %q", resp["message"], "invalid or expired refresh token")
	}
}

func TestRefreshHandler_invalidJSON(t *testing.T) {
	t.Parallel()

	h := newRefreshHandler(&mockRefreshTokenRepo{})
	req := httptest.NewRequest(http.MethodPost, "/users/refresh", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
