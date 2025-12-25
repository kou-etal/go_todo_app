package create

import dtask "github.com/kou-etal/go_todo_app/internal/domain/task"

type Command struct {
	//id作成はdomainの責務
	Title       string
	Description string
	DueDate     dtask.DueOption //nilは期限なし。ゆえにポインタ、いや今は期限必須で。
}

type Result struct {
	ID string
}
