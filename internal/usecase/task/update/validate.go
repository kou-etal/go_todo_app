package update

import (
	"errors"
	"strings"
)

var (
	ErrInvalidID        = errors.New("invalid id")
	ErrInvalidVersion   = errors.New("invalid version")
	ErrNoFieldsToUpdate = errors.New("no fields to update")

	ErrInvalidTitle       = errors.New("invalid title")
	ErrInvalidDescription = errors.New("invalid description")
	ErrInvalidDueOption   = errors.New("invalid due_option")
)

const (
	//ドメイン20文字制限やけど巨大なデータ防ぐためにここである程度制限する。DOS
	maxTitleBytes       = 1024
	maxDescriptionBytes = 16384
)

func normalize(cmd Command) (Command, error) {
	//commandはqueryみたいにhttp query usecase queryにはしない。関数一つでnormaliazation
	if strings.TrimSpace(cmd.ID) == "" {
		return Command{}, ErrInvalidID
	}
	if cmd.Version == 0 {
		return Command{}, ErrInvalidVersion
	}
	if cmd.Title == nil && cmd.Description == nil && cmd.DueDate == nil {
		return Command{}, ErrNoFieldsToUpdate
	} //これ別にフィールド更新なしも可にしてもいいかも

	var (
		title *string
		desc  *string
		due   *int
	) //ifで定義するとスコープ問題。

	if cmd.Title != nil {
		t := strings.TrimSpace(*cmd.Title)
		if len(t) > maxTitleBytes {
			return Command{}, ErrInvalidTitle
		}
		title = &t
	}

	if cmd.Description != nil {
		d := strings.TrimSpace(*cmd.Description)
		if len(d) > maxDescriptionBytes {
			return Command{}, ErrInvalidDescription
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
