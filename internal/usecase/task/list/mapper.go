package list

import (
	"time"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

// 仮にここが変換だけでなく何らかのロジック含んでたらusecaseでテスト
// 現時点ではただの変換ゆえにテストいらん
// Go は 短いスコープでは短い名前
// func mapTaskToItem(taskEntity *dtask.Task) Item スコープが1行なのに長い
func mapTaskToItem(t *dtask.Task) Item {
	duePtr := toTimePtr(t.DueDate().Value())
	return Item{
		ID:          t.ID().Value(),
		Title:       t.Title().Value(),
		Description: t.Description().Value(),
		Status:      t.Status().Value(),
		DueDate:     duePtr,
		//CreatedAt: t.CreatedAt(),
		//これらも使わない。フロントでいつ作られたか表示したいならば使う。cursorでのsortはDBから与えてるから別。
		//cursor使わずフロントでsort定義するなら使うけど良くない
		//UpdatedAt: t.UpdatedAt(),
		//Version:   int64(t.Version()),
		//これは一覧から直接更新APIをたたかないならいらない
	}
}

func toTimePtr(v time.Time) *time.Time {
	x := v
	return &x
} //domainは期限なしを許すルール DTOは期限なしを null で返す表現 ゆえにptrTimeはdomainではなくusecaseで行う
/*func ptrTime(v time.Time) *time.Time {
	return &v
} ポインタ事故回避*/
