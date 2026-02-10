package taskevent

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kou-etal/go_todo_app/internal/domain/task"
	"github.com/kou-etal/go_todo_app/internal/domain/user"
)

type EventID string

func (id EventID) Value() string {
	return string(id)
}

type EventType string //一回宣言の記法

const (
	EventCreated EventType = "created"
	EventUpdated EventType = "updated"
	EventDeleted EventType = "deleted"
)

func ParseEventType(v string) (EventType, error) {
	switch v {
	case string(EventCreated): //stringに合わせて比べる
		return EventCreated, nil
	case string(EventUpdated):
		return EventUpdated, nil
	case string(EventDeleted):
		return EventDeleted, nil
	default:
		return "", errors.New("invalid event type")
	}
}

func (t EventType) Value() string {
	return string(t)
}

type CreatedPayload struct{} //現時点ではcreatedは何も送らない
type UpdatedFields string    //これでtypoも防げる。
const (
	FieldTitle       UpdatedFields = "title"
	FieldDescription UpdatedFields = "description"
	FieldDueDate     UpdatedFields = "due_date"
)

type UpdatedPayload struct {
	Fields []UpdatedFields //1-3の可変長で受け取りたい。これ公開する必要あり。これstringでとるとtype許す。->型作る。
}
type DeletedPayload struct {
	DateLeft int
}
type RequestID string //domainはstring使わずに型でとじる原則

func (id RequestID) Value() string { return string(id) }

type TaskEvent struct {
	id         EventID
	eventType  EventType
	userID     user.UserID
	taskID     task.TaskID
	requestID  RequestID
	occurredAt time.Time
	payload    any
}

// usecaseごとにdonstructer分けることでanyを安全に扱う。
func NewCreatedEvent(
	userID user.UserID,
	taskID task.TaskID,
	now time.Time,
	payload CreatedPayload, //,忘れるミス
) *TaskEvent {
	n := normalizeTime(now) //domain責務

	return &TaskEvent{
		id:         EventID(uuid.New().String()),
		eventType:  EventCreated,
		userID:     userID,
		taskID:     taskID,
		occurredAt: n,
		payload:    payload,
	}
}
func NewDeletedEvent(
	userID user.UserID,
	taskID task.TaskID,
	now time.Time,
	payload DeletedPayload,
) *TaskEvent {
	n := normalizeTime(now)

	return &TaskEvent{
		id:         EventID(uuid.New().String()),
		eventType:  EventDeleted,
		userID:     userID,
		taskID:     taskID,
		occurredAt: n,
		payload:    payload,
	}
}
func NewUpdatedEvent(
	userID user.UserID,
	taskID task.TaskID,
	now time.Time,
	payload UpdatedPayload,
) *TaskEvent {
	n := normalizeTime(now)

	return &TaskEvent{
		id:         EventID(uuid.New().String()),
		eventType:  EventUpdated,
		userID:     userID,
		taskID:     taskID,
		occurredAt: n,
		payload:    payload,
	}
}

func normalizeTime(t time.Time) time.Time {
	return t.UTC().Truncate(time.Second)
} //TODO:これ共有にしたほうがいい
