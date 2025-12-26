package create

//ビジネスロジックで扱わない異常を省く

import (
	"errors"
	"strings"
)

var (
	//これはdomainを作る以前の問題。domainに持ち込まない。
	//エラーの名前抽象的やけど
	ErrInvalidTitle       = errors.New("invalid title")
	ErrInvalidDescription = errors.New("invalid description")
	ErrInvalidDueOption   = errors.New("invalid due_option")
) //ここでエラー定義するのどうなん。ファイル作ったほうがいいか

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
func isValidDueOption(d int) bool {
	switch d {
	case 7, 14, 21, 30:
		return true
	default:
		return false
	}
}

/*
func NewDueOption(t int) (dtask.DueOption, error) {
	switch t {
	case 7:
		return dtask.Due7Days, nil
	case 14:
		return dtask.Due14Days, nil
	case 21:
		return dtask.Due21Days, nil
	case 30:
		return dtask.Due30Days, nil
	default:
		return 0, ErrInvalidDueOption //これって0で返していいん
	}

}*/

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
