package create

//ビジネスロジックで扱わない異常を省く

import (
	"errors"
	"strings"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

var (
	//これはdomainを作る以前の問題。domainに持ち込まない。
	//エラーの名前抽象的やけど
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
	title := strings.TrimSpace(cmd.Title)
	if title == "" {
		return Command{}, ErrInvalidTitle
	}
	if len(title) > maxTitleBytes {
		return Command{}, ErrInvalidTitle
	}

	desc := strings.TrimSpace(cmd.Description)
	if len(desc) > maxDescriptionBytes {
		return Command{}, ErrInvalidDescription
	}

	//UTCとtruncateの責務はdomain。http入力、resultは選択肢

	if !isValidDueOption(cmd.DueDate) {
		return Command{}, ErrInvalidDueOption
	}

	return Command{
		Title:       title,
		Description: desc,
		DueDate:     cmd.DueDate,
	}, nil

}
func isValidDueOption(d dtask.DueOption) bool {
	switch d {
	case dtask.Due7Days, dtask.Due14Days, dtask.Due21Days, dtask.Due30Days:
		return true
	default:
		return false
	}
}

/*
関数にしなかった場合何もしないswitch caseを書くことになる。
due :=cmd.DueOption
	switch due {
	case dtask.Due7Days, dtask.Due14Days, dtask.Due21Days, dtask.Due30Days:

	default:
		return Command{}, ErrInvalidDueOption
	}

	return Command{
		Title:       title,
		Description: desc,
		DueOption:   due,
	}, nil*/
