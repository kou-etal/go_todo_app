package taskrepo

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

type stubQueryerExecer struct {
	gotSQL  string
	gotArgs []any

	records []TaskRecord
	err     error
}

func (s *stubQueryerExecer) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	s.gotSQL = query
	s.gotArgs = args

	if s.err != nil {
		return s.err
	}

	p, ok := dest.(*[]TaskRecord) //型アサーション
	if !ok {
		return errors.New("dest is not *[]TaskRecord")
	}
	*p = append((*p)[:0], s.records...) //stubで返すではなくpに記述
	//(*p)[:0]は中身を空にしたスライス、append先として再利用するための定番実装
	return nil
}
func (s *stubQueryerExecer) GetContext(
	ctx context.Context,
	dest any,
	query string,
	args ...any,
) error {
	return errors.New("stubQueryer.GetContext : unexpected call")
}

func (s *stubQueryerExecer) QueryxContext(
	ctx context.Context,
	query string,
	args ...any,
) (*sqlx.Rows, error) {
	return nil, errors.New("stubQueryer.QueryxContext: unexpected call")
}
func (s *stubQueryerExecer) QueryRowxContext(
	ctx context.Context,
	query string,
	args ...any,
) *sqlx.Row {
	panic("stubQueryer.QueryRowxContext: unexpected call")
}
func (s *stubQueryerExecer) PreparexContext(
	ctx context.Context,
	query string,
) (*sqlx.Stmt, error) {
	return nil, errors.New("stubQueryer.PreparexContext: unexpected call")
}
func (s *stubQueryerExecer) ExecContext(
	ctx context.Context,
	query string,
	args ...any,
) (sql.Result, error) {
	return nil, errors.New("stubQueryerExecuer.ExecContext: unexpected call")
}

func (s *stubQueryerExecer) NamedExecContext(
	ctx context.Context,
	query string,
	arg any,
) (sql.Result, error) {
	return nil, errors.New("stubQueryerExecuer.NamedExecContext: unexpected call")
}

// TaskRecord作成のヘルパー。
// Goの慣習mustXxx=失敗してもエラー拾わない　テストデータ作成はエラー拾ったとこで意味ない。
func mustRecord(id string, created time.Time, due time.Time) TaskRecord {
	return TaskRecord{
		ID:          id,
		Title:       "t",
		Description: "d",
		Status:      "todo",
		DueDate:     due,
		Created:     created,
		Updated:     created,
		Version:     1,
	}
}

// TODO:buildnextの返り値テスト、sort:duedateのテストはいる。
func TestRepository_List_defaultLimitAndSort(t *testing.T) {
	t.Parallel()

	qe := &stubQueryerExecer{}
	repo := NewRepository(qe)

	_, _, err := repo.List(context.Background(), dtask.ListQuery{
		Limit:  50,
		Sort:   "created",
		Cursor: nil,
	})
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}

	if !strings.Contains(qe.gotSQL, "ORDER BY created_at DESC, id DESC") {
		t.Fatalf("sql missing created order:\n%s", qe.gotSQL)
	}

	//dbLimit=q.Limit+1=51がargsの最後に入る
	if len(qe.gotArgs) == 0 {
		t.Fatalf("args is empty")
	}
	last := qe.gotArgs[len(qe.gotArgs)-1]
	if last != 51 {
		t.Fatalf("limit arg = %v, want 51", last)
	}
} //これはすべてが正しい場合のテスト。メリット薄いから消してもいい。

func TestRepository_List_createdCursor_buildsWhereAndArgs(t *testing.T) {
	t.Parallel()

	qe := &stubQueryerExecer{}
	repo := NewRepository(qe)

	cur := &dtask.ListCursor{
		Created: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		ID:      dtask.NewTaskID(),
	}

	_, _, err := repo.List(context.Background(), dtask.ListQuery{
		Limit:  10,
		Sort:   dtask.SortCreated,
		Cursor: cur,
	})
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}

	if !strings.Contains(qe.gotSQL, "WHERE (created_at, id) < (?, ?)") {
		t.Fatalf("sql missing created cursor where:\n%s", qe.gotSQL)
	}
	//created,id,limitの三つ。詳細確認はしない。
	//TODO:と思ったけどこれはrepoの責務だから、もう一段だけ強くして詳細確認するべきかも
	//qr.gotArgs[0] が cur.Created、qr.gotArgs[1] が cur.ID.Value()、qr.gotArgs[2] が 11
	if len(qe.gotArgs) != 3 {
		t.Fatalf("args len = %d, want 3 (created_at,id,limit)", len(qe.gotArgs))
	}
}

/*func TestRepository_List_invalidSort_returnsErr(t *testing.T) {
	t.Parallel()

	qr := &stubQueryer{}
	repo := NewRepository(qr)

	_, _, err := repo.List(context.Background(), dtask.ListQuery{
		Limit:  10,
		Sort:   "___invalid___",
		Cursor: nil,
	})
	if !errors.Is(err, dtask.ErrInvalidSort) {
		t.Fatalf("err = %v, want ErrInvalidSort", err)
	}
}
*/
//TODO:ここに異常なクエリ来ることはあり得ない。それがusecaseの契約。保険のクエリバリデーションまでテストするべきか。するべきらしい。
func TestRepository_List_hasNext_trimsAndReturnsNextCursor(t *testing.T) {
	t.Parallel()

	created1 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	created2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	created3 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Limit=2でrecords=3を返してhasNextを発火させる
	qr := &stubQueryerExecer{
		records: []TaskRecord{
			mustRecord("id-3", created1, time.Time{}),
			mustRecord("id-2", created2, time.Time{}),
			mustRecord("id-1", created3, time.Time{}),
		},
	}
	repo := NewRepository(qr)

	tasks, next, err := repo.List(context.Background(), dtask.ListQuery{
		Limit:  2,
		Sort:   dtask.SortCreated,
		Cursor: nil,
	})
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("tasks len = %d, want 2 (trimmed)", len(tasks))
	}
	if next == nil {
		t.Fatalf("next cursor should not be nil when hasNext")
	}
	//RecordToEntityも巻き込んでる。もしRecordToEntityががっつりロジック持ってるならば分けてテストあるいは固定値。
	// 次カーソルは「返したtasksの最後」を使う
	if !next.Created.Equal(tasks[1].CreatedAt()) {
		t.Fatalf("next.Created = %v, want %v", next.Created, tasks[1].CreatedAt())
	}
	if next.ID != tasks[1].ID() {
		t.Fatalf("next.ID = %v, want %v", next.ID, tasks[1].ID())
	}
}
