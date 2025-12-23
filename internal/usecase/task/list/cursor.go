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
		return dtask.ListCursor{}, dtask.ErrInvalidCursor
	}
	//cursor token は機密でなくてもログに残さない、適当にデータをlogに出力しない。
	//入力が不正

	var c dtask.ListCursor
	if err := json.Unmarshal(b, &c); err != nil {
		return dtask.ListCursor{}, dtask.ErrInvalidCursor
	}
	return c, nil
}
