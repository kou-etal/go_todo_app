package list

import (
	"encoding/base64"
	"errors"
	"testing"
	"time"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

func TestEncodeDecode_roundTrip(t *testing.T) {
	t.Parallel()

	original := dtask.ListCursor{
		ID:      "abc-123",
		Created: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		DueDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	encoded, err := encodeCursor(original)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	if encoded == "" {
		t.Fatal("encoded should not be empty")
	}

	decoded, err := decodeCursor(encoded)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if decoded.ID != original.ID {
		t.Fatalf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if !decoded.Created.Equal(original.Created) {
		t.Fatalf("Created = %v, want %v", decoded.Created, original.Created)
	}
}

func TestDecodeCursor_invalidBase64(t *testing.T) {
	t.Parallel()

	_, err := decodeCursor("!!!invalid!!!")
	if !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("err = %v, want ErrInvalidCursor", err)
	}
}

func TestDecodeCursor_invalidJSON(t *testing.T) {
	t.Parallel()

	encoded := base64.RawURLEncoding.EncodeToString([]byte("not json"))
	_, err := decodeCursor(encoded)
	if !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("err = %v, want ErrInvalidCursor", err)
	}
}
