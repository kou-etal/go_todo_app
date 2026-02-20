package create

import (
	"errors"
	"strings"
)

var (
	ErrEmptyTitle         = errors.New("empty title")
	ErrTitleTooLong       = errors.New("title too long")
	ErrEmptyDescription   = errors.New("empty description")
	ErrDescriptionTooLong = errors.New("description too long")
	ErrInvalidDueOption   = errors.New("invalid due_option")
)

const (
	//ドメイン20文字制限やけど巨大なデータ防ぐためにここである程度制限する。DOS
	maxTitleBytes       = 1024
	maxDescriptionBytes = 16384
)

func normalize(cmd Command) (Command, error) {

	title := strings.TrimSpace(cmd.Title)
	if title == "" {
		return Command{}, ErrEmptyTitle
	}
	if len(title) > maxTitleBytes {
		return Command{}, ErrTitleTooLong
	}

	desc := strings.TrimSpace(cmd.Description)
	if len(desc) > maxDescriptionBytes {
		return Command{}, ErrDescriptionTooLong
	}

	if !isValidDueOption(cmd.DueDate) {
		return Command{}, ErrInvalidDueOption
	}

	return Command{
		Title:       title,
		Description: desc,
		DueDate:     cmd.DueDate,
	}, nil

}
func isValidDueOption(d int) bool {
	switch d {
	case 7, 14, 21, 30:
		return true
	default:
		return false
	}
}
