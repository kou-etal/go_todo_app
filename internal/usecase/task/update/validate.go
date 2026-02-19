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
	ErrConflict           = errors.New("conflict")
)

const (
	//ドメイン20文字制限やけど巨大なデータ防ぐためにここである程度制限する。DOS
	maxTitleBytes       = 1024
	maxDescriptionBytes = 16384
)

func normalize(cmd Command) (Command, error) {
	//commandはqueryみたいにhttp query usecase queryにはしない。関数一つでnormaliazation
	if strings.TrimSpace(cmd.ID) == "" { //保険。そもそもhandlerを通らない場合もある(gRPC/CLI/バッチ)。
		return Command{}, ErrInvalidID
	}
	if cmd.Version == 0 { //ここも1始まりを固定必須
		return Command{}, ErrInvalidVersion
	}
	if cmd.Title == nil && cmd.Description == nil && cmd.DueDate == nil {
		//これはオーケストレーションのnormalization->VOではなくusecase責務
		return Command{}, ErrNoFieldsToUpdate
	} //これ別にフィールド更新なしも可にしてもいいかも

	var (
		title *string
		desc  *string
		due   *int
	)
	//ifで定義するとスコープ問題。
	/*if cmd.Title != nil {
	    t :=
	}
	title = &tこれは不可 */

	if cmd.Title != nil {
		//title = &strings.TrimSpace(*cmd.Title)これができないから一回tに格納してる。
		t := strings.TrimSpace(*cmd.Title)
		if len(t) > maxTitleBytes {
			return Command{}, ErrTitleTooLong
		}
		title = &t
	}
	/*スタック->関数が終わると消える。ヒープ->関数が終わってもGCガベージコレクタが持つ。
	title = &tでtはローカル変数やから関数終わったら消える。titleはtのアドレス見てるからtitleもバグる。
	Goではそうならないようにポインタで格納するとき、その変数が関数から出る場合ヒープに格納する。
	これが多くなると動作が遅くなる。*/

	if cmd.Description != nil {
		d := strings.TrimSpace(*cmd.Description)
		if len(d) > maxDescriptionBytes {
			return Command{}, ErrDescriptionTooLong
		}
		desc = &d
	}

	if cmd.DueDate != nil { //dueoptionはdomainではなくusecaseの関心。まあ練習のために制限しただけで実際7, 14, 21, 30はUX悪すぎる。
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
