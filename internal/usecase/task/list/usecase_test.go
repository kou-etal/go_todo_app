package list

import (
	"context"
	"testing"
	"time"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

type stubRepo struct {
	called   int
	gotQuery dtask.ListQuery

	tasks []*dtask.Task
	next  *dtask.ListCursor
	err   error
}

func (s *stubRepo) List(ctx context.Context, q dtask.ListQuery) ([]*dtask.Task, *dtask.ListCursor, error) {
	s.called++
	s.gotQuery = q
	return s.tasks, s.next, s.err
}
func (s *stubRepo) Store(ctx context.Context, t *dtask.Task) error {
	return nil //TODO:テスト
}

func (s *stubRepo) Update(ctx context.Context, t *dtask.Task) error {
	return nil
}
func (s *stubRepo) FindByID(ctx context.Context, id dtask.TaskID) (*dtask.Task, error) {
	return nil, nil
}
func (s *stubRepo) Delete(ctx context.Context, id dtask.TaskID, version uint64) error

var _ dtask.TaskRepository = (*stubRepo)(nil)

// テストは相互作用ないならCのテストいらない。一つずつでいい。
// 正規化一つも通らないケースはメリット薄い。
func TestUsecase_Do_callsRepoWithNormalizedLimit(t *testing.T) {
	t.Parallel()

	repo := &stubRepo{
		tasks: nil,
		next:  nil,
	}
	u := New(repo)

	//sort/limit/cursor全部成功する値、limit正規化
	q := Query{
		Limit:  0,
		Sort:   "created",
		Cursor: "",
	}

	_, err := u.Do(context.Background(), q)
	if err != nil {
		t.Fatalf("Do() unexpected error: %v", err)
	}
	if repo.called != 1 {
		t.Fatalf("repo.List call count = %d, want 1", repo.called)
	}

	if repo.gotQuery.Limit != 50 {
		t.Fatalf("repo query limit = %d, want 50", repo.gotQuery.Limit)
	}

	if repo.gotQuery.Sort != dtask.SortCreated {
		t.Fatalf("repo query sort is empty; want non-empty(default applied) or set explicit Sort in test")
	}

	if repo.gotQuery.Cursor != nil {
		t.Fatalf("repo query cursor = %#v, want nil", repo.gotQuery.Cursor)
	}
}
func TestUsecase_Do_callsRepoWithNormalizedSort(t *testing.T) {
	t.Parallel()

	repo := &stubRepo{
		tasks: nil,
		next:  nil,
	} //=repo:&stubrepo{}
	u := New(repo)

	//sort/limit/cursor全部成功する値、sort正規化
	q := Query{
		Limit:  5,
		Sort:   "",
		Cursor: "",
	}

	_, err := u.Do(context.Background(), q)
	if err != nil {
		t.Fatalf("Do() unexpected error: %v", err)
	}
	if repo.called != 1 {
		t.Fatalf("repo.List call count = %d, want 1", repo.called)
	}

	if repo.gotQuery.Limit != 5 {
		t.Fatalf("repo query limit = %d, want 5", repo.gotQuery.Limit)
	}

	if repo.gotQuery.Sort != dtask.SortCreated {
		t.Fatalf("repo query sort is empty; want non-empty(default applied) or set explicit Sort in test")
	}

	if repo.gotQuery.Cursor != nil {
		t.Fatalf("repo query cursor = %#v, want nil", repo.gotQuery.Cursor)
	}
}

func TestUsecase_Do_invalidLimit_doesNotCallRepo(t *testing.T) {
	t.Parallel()

	repo := &stubRepo{}
	u := New(repo)

	q := Query{
		Limit:  -1,
		Sort:   "created",
		Cursor: "",
	}

	_, err := u.Do(context.Background(), q)
	if err == nil {
		t.Fatalf("Do() expected error, got nil")
	}
	if repo.called != 0 {
		t.Fatalf("repo.List should not be called, but called %d times", repo.called)
	}
}

func TestUsecase_Do_invalidSort_doesNotCallRepo(t *testing.T) {
	t.Parallel()

	repo := &stubRepo{}
	u := New(repo)

	q := Query{
		Limit:  10,
		Sort:   "___invalid___",
		Cursor: "",
	}

	_, err := u.Do(context.Background(), q)
	if err == nil {
		t.Fatalf("Do() expected error, got nil")
	}
	if repo.called != 0 {
		t.Fatalf("repo.List should not be called, but called %d times", repo.called)
	}
}

func TestUsecase_Do_invalidCursor_doesNotCallRepo(t *testing.T) {
	t.Parallel()

	repo := &stubRepo{}
	u := New(repo)

	q := Query{
		Limit:  10,
		Sort:   "",
		Cursor: "broken-cursor-string",
	}

	_, err := u.Do(context.Background(), q)
	if err == nil {
		t.Fatalf("Do() expected error, got nil")
	}
	if repo.called != 0 {
		t.Fatalf("repo.List should not be called, but called %d times", repo.called)
	}
}

func TestUsecase_Do_Encodedecode_Callrepo(t *testing.T) {
	t.Parallel()

	c := dtask.ListCursor{
		Created:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		DueIsNull: true,
		ID:        dtask.NewTaskID(),
	}
	token, err := encodeCursor(c)
	if err != nil {
		t.Fatalf("encodeCursor() error: %v", err)
	}
	repo := &stubRepo{}
	u := New(repo)

	q := Query{
		Limit:  10,
		Sort:   "created",
		Cursor: token,
	}

	_, err = u.Do(context.Background(), q)
	if err != nil {
		t.Fatalf("Do() unexpected error: %v", err)
	}
	if repo.called != 1 {
		t.Fatalf("repo.List call count = %d, want 1", repo.called)
	}
	if repo.gotQuery.Cursor == nil {
		t.Fatalf("repo query cursor is nil, want non-nil")
	}

	/*if repo.gotQuery.cursor != c{
		t.Fatalf(": %v", err)
	}*/
} //TODO:これencode decode含めることによって2つともバグってた場合通る可能性あり。decodeは固有値を持たせて試すべきか

func TestUsecase_Do_nextCursorGenerated(t *testing.T) {
	t.Parallel()

	next := dtask.ListCursor{
		Created:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		DueIsNull: true,
		ID:        dtask.NewTaskID(),
	} //next定義しないとDoはnextcursor作らない。これは出力のcursor

	repo := &stubRepo{
		tasks: nil,
		next:  &next,
	}
	u := New(repo)

	res, err := u.Do(context.Background(), Query{
		Limit:  10,
		Sort:   "created",
		Cursor: "", //入力のcursor
	})
	if err != nil {
		t.Fatalf("Do() unexpected error: %v", err)
	}
	if res.NextCursor == "" {
		t.Fatalf("NextCursor should not be empty when repo returns next")
	} //cursorの中身まではテストしてない。それはcursor.goのテスト責務
}
