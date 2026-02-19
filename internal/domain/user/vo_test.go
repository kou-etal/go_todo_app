package user

import (
	"errors"
	"strings"
	"testing"
)

// --- stub ---

type stubHasher struct {
	hash string
	err  error
}

func (s stubHasher) Hash(_ string) (string, error) { return s.hash, s.err }
func (s stubHasher) Compare(_, _ string) error      { return nil }

// --- tests ---

func TestNewUserEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantValue string
		wantErr   error
	}{
		{"valid", "user@example.com", "user@example.com", nil},
		{"uppercase_normalized", "USER@EXAMPLE.COM", "user@example.com", nil},
		{"trimmed", "  user@example.com  ", "user@example.com", nil},
		{"empty", "", "", ErrEmptyEmail},
		{"whitespace_only", "   ", "", ErrEmptyEmail},
		{"no_at", "not-an-email", "", ErrInvalidEmailFormat},
		{"too_long", strings.Repeat("a", 243) + "@example.com", "", ErrEmailTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewUserEmail(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Value() != tt.wantValue {
				t.Fatalf("value = %q, want %q", got.Value(), tt.wantValue)
			}
		})
	}
}

func TestNewUserPasswordFromPlain(t *testing.T) {
	t.Parallel()

	okHasher := stubHasher{hash: "$2a$10$hashed"}
	errHasher := stubHasher{err: errors.New("hash failed")}

	tests := []struct {
		name    string
		plain   string
		hasher  PasswordHasher
		wantErr error
	}{
		{"valid_12_chars", "password1234", okHasher, nil},
		{"empty", "", okHasher, ErrEmptyPassword},
		{"whitespace_only", "   ", okHasher, ErrEmptyPassword},
		{"leading_space", " password1234", okHasher, ErrPasswordHasLeadingOrTrailingSpace},
		{"trailing_space", "password1234 ", okHasher, ErrPasswordHasLeadingOrTrailingSpace},
		{"too_short_11", "password123", okHasher, ErrPasswordTooShort},
		{"too_long_73_bytes", strings.Repeat("a", 73), okHasher, ErrPasswordTooLong},
		{"exactly_72_bytes", strings.Repeat("a", 72), okHasher, nil},
		{"hasher_error", "password1234", errHasher, errHasher.err},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewUserPasswordFromPlain(tt.plain, tt.hasher)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Hash() == "" {
				t.Fatal("hash should not be empty")
			}
		})
	}
}

func TestNewUserName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantValue string
		wantErr   error
	}{
		{"valid", "alice", "alice", nil},
		{"trimmed", "  alice  ", "alice", nil},
		{"empty", "", "", ErrEmptyName},
		{"whitespace_only", "   ", "", ErrEmptyName},
		{"exactly_20_runes", strings.Repeat("a", 20), strings.Repeat("a", 20), nil},
		{"21_runes_too_long", strings.Repeat("a", 21), "", ErrNameTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewUserName(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Value() != tt.wantValue {
				t.Fatalf("value = %q, want %q", got.Value(), tt.wantValue)
			}
		})
	}
}
