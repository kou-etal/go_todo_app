package task

import (
	"errors"
	"time"
)

// aggregate:=単体でフィールド変えられたくない。ルールの強さ。ルールを破られると壊れる。テーブル単位で考える。
// TaskはstatusやdueDateに遷移ルールあり、ゆえにaggregate
// ゆえに勝手にどこからでもフィールド変えられたくないから更新系をChangeでレシーバにする。
// userprofileとかは別に強いルールない。aggregateじゃないこと多い。
// 何でもかんでもaggregateにしてchangeメソッド作ると良くない。
type Task struct {
	id          TaskID //学習用メモ　ここでDBとのマッピング定義するとdomainにDB都合入って良くない。
	title       TaskTitle
	description TaskDescription
	status      TaskStatus
	dueDate     DueDate
	createdAt   time.Time
	updatedAt   time.Time
	version     uint64
}

type Tasks []*Task

//TODO:Tasksを定義するならば定義に意味を持たせるようなメソッドをタスクに持たせる。ただ短いからTasksは微妙。

func (t *Task) ID() TaskID                   { return t.id }
func (t *Task) Title() TaskTitle             { return t.title }
func (t *Task) Description() TaskDescription { return t.description }
func (t *Task) Status() TaskStatus           { return t.status }
func (t *Task) DueDate() DueDate             { return t.dueDate }
func (t *Task) CreatedAt() time.Time         { return t.createdAt }
func (t *Task) UpdatedAt() time.Time         { return t.updatedAt }
func (t *Task) Version() uint64              { return t.version }

func (t *Task) ChangeTitle(newTitle TaskTitle, now time.Time) {
	t.title = newTitle
	t.updateTime(now)
}
func (t *Task) ChangeDescription(newDesc TaskDescription, now time.Time) {
	t.description = newDesc
	t.updateTime(now)
}
func (t *Task) ChangeDueDate(newDue DueDate, now time.Time) {
	t.dueDate = newDue
	t.updateTime(now)
} //TODO:遷移ルール定義してもいい
//過去禁止はfactoryで定義

func (t *Task) ChangeStatus(next TaskStatus, now time.Time) error {
	//TODO:状態遷移ルールが弱い.canChangeStatusではなくここで定義する。
	if t.status == StatusDone {
		return errors.New("tmp")
	}

	t.status = next
	t.updateTime(now)
	return nil
}

func (t *Task) Reschedule(newDue DueDate, now time.Time) {
	//TODO:これも状態遷移系。
	t.dueDate = newDue
	t.updateTime(now)
}

// MarkDoneに意味を持たせるならばchangestatusと分離
func (t *Task) MarkDone(now time.Time) error {
	return t.ChangeStatus(StatusDone, now)
}

func (t *Task) updateTime(now time.Time) {
	n := normalizeTime(now)
	t.updatedAt = n
}

func normalizeTime(t time.Time) time.Time {
	return t.UTC().Truncate(time.Second)
}

/*func (t *Task) updateTime(now time.Time) {
	//TODO:versionの操作はrepo層に任せる
	n := normalizeTime(now)
	if n.After(t.updatedAt) {
		t.updatedAt = n
		t.version++
	}
}*/
//これやとフィールドごとにvesion増えておかしい。
//repoで更新完了したら+にするべき
