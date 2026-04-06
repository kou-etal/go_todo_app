package refresh

import (
	"errors"
	"strings"
)

var (
	ErrEmptyRefreshToken  = errors.New("refresh token is required")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

func normalize(cmd Command) (Command, error) {
	cmd.RefreshToken = strings.TrimSpace(cmd.RefreshToken)

	if cmd.RefreshToken == "" {
		return Command{}, ErrEmptyRefreshToken
	}

	return cmd, nil
}
