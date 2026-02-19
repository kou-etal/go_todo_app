package register

import (
	"context"
	"errors"
	"testing"
	"time"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

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

// --- mock deps ---

type mockRegisterDeps struct {
	userRepo        duser.UserRepository
	emailVerifyRepo dverify.EmailVerificationRepository
}

func (m *mockRegisterDeps) UserRepo() duser.UserRepository                     { return m.userRepo }
func (m *mockRegisterDeps) EmailVerifyRepo() dverify.EmailVerificationRepository { return m.emailVerifyRepo }

// --- mock user repo ---

type mockUserRepo struct {
	storeErr    error
	storeCalled bool
}

func (m *mockUserRepo) Store(_ context.Context, _ *duser.User) error {
	m.storeCalled = true
	return m.storeErr
}

// --- mock email verify repo ---

type mockEmailVerifyRepo struct {
	insertErr    error
	insertCalled bool
}

func (m *mockEmailVerifyRepo) Insert(_ context.Context, _ *dverify.EmailVerificationToken) error {
	m.insertCalled = true
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

// --- mock clocker ---

type mockClocker struct {
	now time.Time
}

func (m mockClocker) Now() time.Time { return m.now }

// --- helpers ---

func validCommand() Command {
	return Command{
		Email:    "test@example.com",
		Password: "password1234",
		UserName: "alice",
	}
}

func setupUsecase(userRepo *mockUserRepo, emailRepo *mockEmailVerifyRepo) *Usecase {
	deps := &mockRegisterDeps{userRepo: userRepo, emailVerifyRepo: emailRepo}
	runner := &mockRunner{deps: deps}
	clk := mockClocker{now: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)}
	return New(
		runner, clk,
		stubHasher{hash: "$2a$10$hashed"},
		stubTokenGenerator{token: "plain-token"},
		stubTokenHasher{hash: "sha256-hash"},
	)
}

// --- tests ---

func TestDo_happyPath(t *testing.T) {
	t.Parallel()

	userRepo := &mockUserRepo{}
	emailRepo := &mockEmailVerifyRepo{}
	uc := setupUsecase(userRepo, emailRepo)

	res, err := uc.Do(context.Background(), validCommand())
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	if res.UserID == "" {
		t.Fatal("result.UserID should not be empty")
	}
	if !userRepo.storeCalled {
		t.Fatal("UserRepo.Store was not called")
	}
	if !emailRepo.insertCalled {
		t.Fatal("EmailVerifyRepo.Insert was not called")
	}
}

func TestDo_emptyEmail(t *testing.T) {
	t.Parallel()

	uc := setupUsecase(&mockUserRepo{}, &mockEmailVerifyRepo{})
	cmd := validCommand()
	cmd.Email = ""

	_, err := uc.Do(context.Background(), cmd)
	if !errors.Is(err, ErrEmptyEmail) {
		t.Fatalf("err = %v, want ErrEmptyEmail", err)
	}
}

func TestDo_emptyPassword(t *testing.T) {
	t.Parallel()

	uc := setupUsecase(&mockUserRepo{}, &mockEmailVerifyRepo{})
	cmd := validCommand()
	cmd.Password = ""

	_, err := uc.Do(context.Background(), cmd)
	if !errors.Is(err, ErrEmptyPassword) {
		t.Fatalf("err = %v, want ErrEmptyPassword", err)
	}
}

func TestDo_emptyUserName(t *testing.T) {
	t.Parallel()

	uc := setupUsecase(&mockUserRepo{}, &mockEmailVerifyRepo{})
	cmd := validCommand()
	cmd.UserName = ""

	_, err := uc.Do(context.Background(), cmd)
	if !errors.Is(err, ErrEmptyUserName) {
		t.Fatalf("err = %v, want ErrEmptyUserName", err)
	}
}

func TestDo_invalidEmailFormat(t *testing.T) {
	t.Parallel()

	uc := setupUsecase(&mockUserRepo{}, &mockEmailVerifyRepo{})
	cmd := validCommand()
	cmd.Email = "not-an-email"

	_, err := uc.Do(context.Background(), cmd)
	if !errors.Is(err, ErrInvalidEmailFormat) {
		t.Fatalf("err = %v, want ErrInvalidEmailFormat", err)
	}
}

func TestDo_passwordTooShort(t *testing.T) {
	t.Parallel()

	uc := setupUsecase(&mockUserRepo{}, &mockEmailVerifyRepo{})
	cmd := validCommand()
	cmd.Password = "short12345a"

	_, err := uc.Do(context.Background(), cmd)
	if !errors.Is(err, ErrPasswordTooShort) {
		t.Fatalf("err = %v, want ErrPasswordTooShort", err)
	}
}

func TestDo_storeConflict(t *testing.T) {
	t.Parallel()

	userRepo := &mockUserRepo{storeErr: duser.ErrConflict}
	emailRepo := &mockEmailVerifyRepo{}
	uc := setupUsecase(userRepo, emailRepo)

	_, err := uc.Do(context.Background(), validCommand())
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}
