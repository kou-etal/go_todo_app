package list

//TODO:これはテスト書く価値そこそこある。cursorの仕様は独立して壊れやすいからcursor単体テストはコスパ良い。
import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

func encodeCursor(c dtask.ListCursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("encode cursor: %w", err)
	}
	//これは起こりえない。fail-fastでpanic　or  errorで伝搬
	//json.Marshal が落ちるのは入力が不正ではないからdtask.ErrInvalidCursorではない
	//それをErrInvalidCursorでかえすと内部バグを400系に誤分類
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func decodeCursor(token string) (dtask.ListCursor, error) {
	b, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return dtask.ListCursor{}, ErrInvalidCursor
	}

	var c dtask.ListCursor
	if err := json.Unmarshal(b, &c); err != nil {
		return dtask.ListCursor{}, ErrInvalidCursor
	}
	return c, nil
}
