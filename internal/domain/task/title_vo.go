package task

import (
	"strings"
	"unicode/utf8"
)

// そもそもなんでstructにしてる。
// type TaskTitle stringの場合別パッケージからvar t TaskTitle = "x"が可能、newtasktitleを通らない不正な値が発生。
// structの場合var t TaskTitle = "x"、TaskTitle{valuse:x}が別パッケージから不可能。
// さらにstructの場合フィールド変更も簡単、メソッドもはやせる。
// idは単なる識別子やからtypeでいい
type TaskTitle struct {
	value string
}

func NewTaskTitle(v string) (TaskTitle, error) {
	if strings.TrimSpace(v) == "" {
		return TaskTitle{}, ErrEmptyTitle
	}
	if utf8.RuneCountInString(v) > 20 {
		return TaskTitle{}, ErrTitleTooLong
	}

	return TaskTitle{value: v}, nil
}

func (t TaskTitle) Value() string {
	return t.value
}
