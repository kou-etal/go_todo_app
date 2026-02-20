package list

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

func encodeCursor(c dtask.ListCursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("encode cursor: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func decodeCursor(token string) (dtask.ListCursor, error) {
	b, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return dtask.ListCursor{}, ErrInvalidCursor
	}

	var c dtask.ListCursor
	if err := json.Unmarshal(b, &c); err != nil {
		return dtask.ListCursor{}, ErrInvalidCursor
	}
	return c, nil
}
