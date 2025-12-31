package remove

import (
	"errors"
	"strings"
)

var (
	ErrInvalidID      = errors.New("invalid id")
	ErrInvalidVersion = errors.New("invalid version")
)

func normalize(cmd Command) (Command, error) {
	//commandはqueryみたいにhttp query usecase queryにはしない。関数一つでnormaliazation
	//deleteはこれだけ
	if strings.TrimSpace(cmd.ID) == "" {
		return Command{}, ErrInvalidID
	}
	if cmd.Version == 0 {
		return Command{}, ErrInvalidVersion
	}

	return Command{
		ID:      cmd.ID,
		Version: cmd.Version,
	}, nil
}
