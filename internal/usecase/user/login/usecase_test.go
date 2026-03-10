package login

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
	deps    usetx.LoginDeps
	withErr error
}

func (m *mockRunner) WithinTx(ctx context.Context, fn func(ctx context.Context, deps usetx.LoginDeps) error) error {
	if m.withErr != nil {
		return m.withErr
	}
	return fn(ctx, m.deps)
}

// --- mock deps ---

type mockLoginDeps struct {
	userRepo         duser.UserRepository
	refreshTokenRepo drefresh.RefreshTokenRepository
}

func (m *mockLoginDeps) UserRepo() duser.UserRepository                { return m.userRepo }
func (m *mockLoginDeps) RefreshTokenRepo() drefresh.RefreshTokenRepository { return m.refreshTokenRepo }

// --- mock user repo ---

type mockUserRepo struct {
	findUser *duser.User
	findErr  error
}

func (m *mockUserRepo) Store(_ context.Context, _ *duser.User) error { return nil }
func (m *mockUserRepo) FindByEmail(_ context.Context, _ duser.UserEmail) (*duser.User, error) {
	return m.findUser, m.findErr
}

// --- mock refresh token repo ---

type mockRefreshRepo struct {
	storeCalled bool
	storeErr    error
}

func (m *mockRefreshRepo) Store(_ context.Context, _ *drefresh.RefreshToken) error {
	m.storeCalled = true
	return m.storeErr
}
func (m *mockRefreshRepo) FindByTokenHashForUpdate(_ context.Context, _ drefresh.TokenHash) (*drefresh.RefreshToken, error) {
	return nil, nil
}
func (m *mockRefreshRepo) Update(_ context.Context, _ *drefresh.RefreshToken) error { return nil }

// --- stubs ---

type stubClocker struct{ now time.Time }

func (c stubClocker) Now() time.Time { return c.now }

type stubPasswordHasher struct {
	hash       string
	hashErr    error
	compareErr error
}

func (s stubPasswordHasher) Hash(_ string) (string, error) { return s.hash, s.hashErr }
func (s stubPasswordHasher) Compare(_, _ string) error      { return s.compareErr }

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

func testUser() *duser.User {
	email, _ := duser.NewUserEmail("test@example.com")
	pw, _ := duser.ReconstructUserPassword("$2a$10$hashed")
	name, _ := duser.NewUserName("alice")
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return duser.ReconstructUser(duser.UserID(testUID), email, pw, name, nil, now, now, 1)
}

func setupUsecase(userRepo *mockUserRepo, refreshRepo *mockRefreshRepo, pwHasher stubPasswordHasher) *Usecase {
	deps := &mockLoginDeps{userRepo: userRepo, refreshTokenRepo: refreshRepo}
	runner := &mockRunner{deps: deps}
	clk := stubClocker{now: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)}
	return New(
		runner,
		clk,
		pwHasher,
		stubTokenGen{token: "access-token-123"},
		stubRefreshHasher{hash: "sha256-hash"},
		stubRefreshGen{token: "refresh-token-456"},
		7*24*time.Hour,
		900,
	)
}

// --- tests ---

func TestDo_happyPath(t *testing.T) {
	t.Parallel()

	userRepo := &mockUserRepo{findUser: testUser()}
	refreshRepo := &mockRefreshRepo{}
	uc := setupUsecase(userRepo, refreshRepo, stubPasswordHasher{hash: "$2a$10$hashed"})

	res, err := uc.Do(context.Background(), Command{
		Email:    "test@example.com",
		Password: "password1234",
	})
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	if res.AccessToken == "" {
		t.Fatal("access token should not be empty")
	}
	if res.RefreshToken == "" {
		t.Fatal("refresh token should not be empty")
	}
	if res.ExpiresIn != 900 {
		t.Fatalf("expires_in = %d, want 900", res.ExpiresIn)
	}
	if !refreshRepo.storeCalled {
		t.Fatal("refresh token Store was not called")
	}
}

func TestDo_emptyEmail(t *testing.T) {
	t.Parallel()

	uc := setupUsecase(&mockUserRepo{}, &mockRefreshRepo{}, stubPasswordHasher{})

	_, err := uc.Do(context.Background(), Command{
		Email:    "",
		Password: "password1234",
	})
	if !errors.Is(err, ErrEmptyEmail) {
		t.Fatalf("err = %v, want ErrEmptyEmail", err)
	}
}

func TestDo_emptyPassword(t *testing.T) {
	t.Parallel()

	uc := setupUsecase(&mockUserRepo{}, &mockRefreshRepo{}, stubPasswordHasher{})

	_, err := uc.Do(context.Background(), Command{
		Email:    "test@example.com",
		Password: "",
	})
	if !errors.Is(err, ErrEmptyPassword) {
		t.Fatalf("err = %v, want ErrEmptyPassword", err)
	}
}

func TestDo_userNotFound(t *testing.T) {
	t.Parallel()

	userRepo := &mockUserRepo{findErr: duser.ErrNotFound}
	uc := setupUsecase(userRepo, &mockRefreshRepo{}, stubPasswordHasher{})

	_, err := uc.Do(context.Background(), Command{
		Email:    "unknown@example.com",
		Password: "password1234",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestDo_wrongPassword(t *testing.T) {
	t.Parallel()

	userRepo := &mockUserRepo{findUser: testUser()}
	pwHasher := stubPasswordHasher{compareErr: errors.New("mismatch")}
	uc := setupUsecase(userRepo, &mockRefreshRepo{}, pwHasher)

	_, err := uc.Do(context.Background(), Command{
		Email:    "test@example.com",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestDo_refreshStoreError(t *testing.T) {
	t.Parallel()

	userRepo := &mockUserRepo{findUser: testUser()}
	storeErr := errors.New("db write failed")
	refreshRepo := &mockRefreshRepo{storeErr: storeErr}
	uc := setupUsecase(userRepo, refreshRepo, stubPasswordHasher{})

	_, err := uc.Do(context.Background(), Command{
		Email:    "test@example.com",
		Password: "password1234",
	})
	if !errors.Is(err, storeErr) {
		t.Fatalf("err = %v, want %v", err, storeErr)
	}
}
