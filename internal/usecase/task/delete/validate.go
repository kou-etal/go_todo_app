package remove

import (
	"errors"
	"strings"
)

var (
	ErrInvalidID      = errors.New("invalid id")
	ErrInvalidVersion = errors.New("invalid version")
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")
)

func normalize(cmd Command) (Command, error) {

	if strings.TrimSpace(cmd.ID) == "" {
		return Command{}, ErrInvalidID
	}
	if cmd.Version == 0 {
		return Command{}, ErrInvalidVersion
	}

	return Command{
		UserID:  cmd.UserID,
		ID:      cmd.ID,
		Version: cmd.Version,
	}, nil
}
