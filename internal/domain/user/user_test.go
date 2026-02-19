package user

import (
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

	email, _ := NewUserEmail("test@example.com")
	pass, _ := NewUserPasswordFromPlain("password1234", stubHasher{hash: "$2a$hashed"})
	name, _ := NewUserName("alice")
	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	u := NewUser(email, pass, name, now)

	if u.ID().Value() == "" {
		t.Fatal("ID should not be empty")
	}
	if u.Version() != 1 {
		t.Fatalf("version = %d, want 1", u.Version())
	}
	if u.EmailVerifiedAt() != nil {
		t.Fatal("emailVerifiedAt should be nil")
	}
	if !u.CreatedAt().Equal(u.UpdatedAt()) {
		t.Fatal("createdAt should equal updatedAt")
	}
}

func TestVerifyEmail(t *testing.T) {
	t.Parallel()

	email, _ := NewUserEmail("test@example.com")
	pass, _ := NewUserPasswordFromPlain("password1234", stubHasher{hash: "$2a$hashed"})
	name, _ := NewUserName("alice")
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	u := NewUser(email, pass, name, created)
	if u.EmailVerifiedAt() != nil {
		t.Fatal("precondition: emailVerifiedAt should be nil")
	}

	verifyTime := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	u.VerifyEmail(verifyTime)

	if u.EmailVerifiedAt() == nil {
		t.Fatal("emailVerifiedAt should not be nil after VerifyEmail")
	}
	if u.EmailVerifiedAt().Location().String() != "UTC" {
		t.Fatalf("emailVerifiedAt location = %v, want UTC", u.EmailVerifiedAt().Location())
	}
}
