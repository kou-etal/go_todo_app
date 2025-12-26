package list

import (
	"context"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

type Usecase struct {
	repo dtask.TaskRepository
}

func New(repo dtask.TaskRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) Do(ctx context.Context, q Query) (Result, error) {
	limit, err := normalizeLimit(q.Limit)
	if err != nil {
		return Result{}, err
	}

	sort, err := normalizeSort(q.Sort)
	if err != nil {
		return Result{}, err
	}

	var cursor *dtask.ListCursor
	if q.Cursor != "" {
		c, err := decodeCursor(q.Cursor)
		if err != nil {
			return Result{}, err
			//decodecursorの定義側で返してる。
			//低レベル関数を複数用途で使うときに呼び出し側でエラー定義
		}
		cursor = &c
	}

	dq := dtask.ListQuery{
		Limit:  limit,
		Sort:   sort,
		Cursor: cursor,
	}

	tasks, next, err := u.repo.List(ctx, dq)
	if err != nil {
		return Result{}, err
	}

	items := make([]Item, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, mapTaskToItem(t))
	}
	//これただのデータ成形やからテストいらん。
	// もしmapperが変換だけでなくロジック持ってたり、ここでitemの総数制限とかロジック含んでたらテスト.

	nextCursor := ""
	//空文字比較は罠
	if next != nil {
		nextCursor, err = encodeCursor(*next)
		if err != nil {
			return Result{}, err
		}
	}

	return Result{
		Items:      items,
		NextCursor: nextCursor,
	}, nil
}
