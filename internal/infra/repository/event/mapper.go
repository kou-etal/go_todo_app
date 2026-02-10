package taskeventrepo

import (
	"encoding/json"
	"fmt"

	taskevent "github.com/kou-etal/go_todo_app/internal/domain/event"
)

// append-onlyやからtoEntity使わない。

// ポインタにするの忘れがち
func toRecord(e *taskevent.TaskEvent) (TaskEventRecord, error) {
	//marshalがtojson
	p, err := json.Marshal(e.Payload()) //ここではまだbyte
	if err != nil {
		return TaskEventRecord{}, fmt.Errorf("marshal payload: %w", err) //これは想定外。wrapする。
	}
	return TaskEventRecord{
		ID:            e.ID().Value(),
		UserID:        e.UserID().Value(),
		TaskID:        e.TaskID().Value(),
		RequestID:     e.RequestID().Value(),
		EventType:     e.EventType().Value(),
		OccurredAt:    e.OccurredAt(),
		EmittedAt:     nil, //ここはrepo実行の際に定義
		SchemaVersion: 1,
		Payload:       json.RawMessage(p), //jsonで使いますよ宣言
	}, nil
}
