package refresh

import (
	"context"
	"errors"
	"testing"
	"time"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	drefresh "github.com/kou-etal/go_todo_app/internal/domain/user/refresh"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

// --- mock runner ---

type mockRunner struct {
	deps    usetx.RefreshDeps
	withErr error
}

func (m *mockRunner) WithinTx(ctx context.Context, fn func(ctx context.Context, deps usetx.RefreshDeps) error) error {
	if m.withErr != nil {
		return m.withErr
	}
	return fn(ctx, m.deps)
}

// --- mock deps ---

type mockRefreshDeps struct {
	refreshTokenRepo drefresh.RefreshTokenRepository
}

func (m *mockRefreshDeps) RefreshTokenRepo() drefresh.RefreshTokenRepository { return m.refreshTokenRepo }

// --- mock refresh token repo ---

type mockRefreshRepo struct {
	findToken   *drefresh.RefreshToken
	findErr     error
	updateErr   error
	storeErr    error
	storeCalled bool
}

func (m *mockRefreshRepo) FindByTokenHashForUpdate(_ context.Context, _ drefresh.TokenHash) (*drefresh.RefreshToken, error) {
	return m.findToken, m.findErr
}
func (m *mockRefreshRepo) Update(_ context.Context, _ *drefresh.RefreshToken) error {
	return m.updateErr
}
func (m *mockRefreshRepo) Store(_ context.Context, _ *drefresh.RefreshToken) error {
	m.storeCalled = true
	return m.storeErr
}

// --- stubs ---

type stubClocker struct{ now time.Time }

func (c stubClocker) Now() time.Time { return c.now }

type stubTokenGen struct {
	token string
	err   error
}

func (s stubTokenGen) GenerateAccessToken(_ string) (string, error) { return s.token, s.err }

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

// --- helpers ---

const testUID = "00000000-0000-0000-0000-000000000001"

func validRefreshToken(now time.Time) *drefresh.RefreshToken {
	tokenHash, _ := drefresh.ReconstructTokenHash("sha256-old-hash")
	return drefresh.ReconstructRefreshToken(
		drefresh.NewTokenID(),
		duser.UserID(testUID),
		tokenHash,
		now.Add(7*24*time.Hour),
		nil,
		now.Add(-1*time.Hour),
	)
}

func expiredRefreshToken(now time.Time) *drefresh.RefreshToken {
	tokenHash, _ := drefresh.ReconstructTokenHash("sha256-old-hash")
	return drefresh.ReconstructRefreshToken(
		drefresh.NewTokenID(),
		duser.UserID(testUID),
		tokenHash,
		now.Add(-1*time.Hour),
		nil,
		now.Add(-48*time.Hour),
	)
}

func revokedRefreshToken(now time.Time) *drefresh.RefreshToken {
	tokenHash, _ := drefresh.ReconstructTokenHash("sha256-old-hash")
	revokedAt := now.Add(-30 * time.Minute)
	return drefresh.ReconstructRefreshToken(
		drefresh.NewTokenID(),
		duser.UserID(testUID),
		tokenHash,
		now.Add(7*24*time.Hour),
		&revokedAt,
		now.Add(-1*time.Hour),
	)
}

func setupUsecase(repo *mockRefreshRepo) *Usecase {
	deps := &mockRefreshDeps{refreshTokenRepo: repo}
	runner := &mockRunner{deps: deps}
	clk := stubClocker{now: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)}
	return New(
		runner,
		clk,
		stubTokenGen{token: "new-access-token"},
		stubRefreshHasher{hash: "sha256-new-hash"},
		stubRefreshGen{token: "new-refresh-token"},
		7*24*time.Hour,
		900,
	)
}

// --- tests ---

func TestDo_happyPath(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	repo := &mockRefreshRepo{findToken: validRefreshToken(now)}
	uc := setupUsecase(repo)

	res, err := uc.Do(context.Background(), Command{RefreshToken: "old-plain-token"})
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	if res.AccessToken != "new-access-token" {
		t.Fatalf("access_token = %q, want %q", res.AccessToken, "new-access-token")
	}
	if res.RefreshToken != "new-refresh-token" {
		t.Fatalf("refresh_token = %q, want %q", res.RefreshToken, "new-refresh-token")
	}
	if res.ExpiresIn != 900 {
		t.Fatalf("expires_in = %d, want 900", res.ExpiresIn)
	}
	if !repo.storeCalled {
		t.Fatal("new refresh token Store was not called")
	}
}

func TestDo_emptyToken(t *testing.T) {
	t.Parallel()

	uc := setupUsecase(&mockRefreshRepo{})

	_, err := uc.Do(context.Background(), Command{RefreshToken: ""})
	if !errors.Is(err, ErrEmptyRefreshToken) {
		t.Fatalf("err = %v, want ErrEmptyRefreshToken", err)
	}
}

func TestDo_tokenNotFound(t *testing.T) {
	t.Parallel()

	repo := &mockRefreshRepo{findErr: drefresh.ErrNotFound}
	uc := setupUsecase(repo)

	_, err := uc.Do(context.Background(), Command{RefreshToken: "unknown-token"})
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("err = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestDo_expiredToken(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	repo := &mockRefreshRepo{findToken: expiredRefreshToken(now)}
	uc := setupUsecase(repo)

	_, err := uc.Do(context.Background(), Command{RefreshToken: "expired-token"})
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("err = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestDo_revokedToken(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	repo := &mockRefreshRepo{findToken: revokedRefreshToken(now)}
	uc := setupUsecase(repo)

	_, err := uc.Do(context.Background(), Command{RefreshToken: "revoked-token"})
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("err = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestDo_storeError(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	storeErr := errors.New("db write failed")
	repo := &mockRefreshRepo{findToken: validRefreshToken(now), storeErr: storeErr}
	uc := setupUsecase(repo)

	_, err := uc.Do(context.Background(), Command{RefreshToken: "old-plain-token"})
	if !errors.Is(err, storeErr) {
		t.Fatalf("err = %v, want %v", err, storeErr)
	}
}
