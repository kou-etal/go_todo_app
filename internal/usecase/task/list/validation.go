package list

import (
	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

// これはusecaseで扱ったからテストいらん。
const (
	defaultLimit = 50
	maxLimit     = 200
)

// これはhttp入力をバリデーションするわけではなく(handler)usecaseの契約を守る(http query)
// limit補正はdomain	repoの責務ではないアプリの都合
// ここで定義するとユースケースがHTTP以外のgRPC/CLI/Jobから使われても挙動一定になる
func normalizeLimit(v int) (int, error) {
	if v == 0 {
		return defaultLimit, nil
	}

	if v < 1 || v > maxLimit {
		return 0, dtask.ErrInvalidLimit
	}
	return v, nil
}

func normalizeSort(v string) (dtask.ListSort, error) {
	switch v {
	case "", "created":
		return dtask.SortCreated, nil
	case "dueDate":
		return dtask.SortDueDate, nil
	default:
		return "", dtask.ErrInvalidSort
	}
}
