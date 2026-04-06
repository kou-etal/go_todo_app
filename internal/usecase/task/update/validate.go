package update

import (
	"errors"
	"strings"
)

var (
	ErrInvalidID          = errors.New("invalid id")
	ErrInvalidVersion     = errors.New("invalid version")
	ErrNoFieldsToUpdate   = errors.New("no fields to update")
	ErrEmptyTitle         = errors.New("empty title")
	ErrTitleTooLong       = errors.New("title too long")
	ErrEmptyDescription   = errors.New("empty description")
	ErrDescriptionTooLong = errors.New("description too long")
	ErrInvalidDueOption   = errors.New("invalid due_option")
	ErrNotFound           = errors.New("not found")
	ErrForbidden          = errors.New("forbidden")
	ErrConflict           = errors.New("conflict")
)

const (
	//ドメイン20文字制限やけど巨大なデータ防ぐためにここである程度制限する。DOS
	maxTitleBytes       = 1024
	maxDescriptionBytes = 16384
)

func normalize(cmd Command) (Command, error) {

	if strings.TrimSpace(cmd.ID) == "" { //保険。そもそもhandlerを通らない場合もある(gRPC/CLI/バッチ)。
		return Command{}, ErrInvalidID
	}
	if cmd.Version == 0 {
		return Command{}, ErrInvalidVersion
	}
	if cmd.Title == nil && cmd.Description == nil && cmd.DueDate == nil {

		return Command{}, ErrNoFieldsToUpdate
	}

	var (
		title *string
		desc  *string
		due   *int
	)

	if cmd.Title != nil {

		t := strings.TrimSpace(*cmd.Title)
		if len(t) > maxTitleBytes {
			return Command{}, ErrTitleTooLong
		}
		title = &t
	}

	if cmd.Description != nil {
		d := strings.TrimSpace(*cmd.Description)
		if len(d) > maxDescriptionBytes {
			return Command{}, ErrDescriptionTooLong
		}
		desc = &d
	}

	if cmd.DueDate != nil {
		if !isValidDueOption(*cmd.DueDate) {
			return Command{}, ErrInvalidDueOption
		}
		d := *cmd.DueDate
		due = &d
	}

	return Command{
		UserID:      cmd.UserID,
		ID:          cmd.ID,
		Version:     cmd.Version,
		Title:       title,
		Description: desc,
		DueDate:     due,
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
