package task

import "errors"

var (
	ErrInvalidSort        = errors.New("invalid sort")
	ErrInvalidCursor      = errors.New("invalid cursor")
	ErrInvalidLimit       = errors.New("invalid limit")
	ErrConflict           = errors.New("taskrepo: conflict")
	ErrEmptyTitle         = errors.New("need title") //入力不正の詳細
	ErrTitleTooLong       = errors.New("too long title")
	ErrEmptyDescription   = errors.New("need description")
	ErrDescriptionTooLong = errors.New("too long description")
	ErrInvalidDueOption   = errors.New("invalid due")
	ErrInvalidID          = errors.New("invalid ID")
	ErrNotFound           = errors.New("not found")
)

/*
istsortとかlistcursorをdomainで定義してるからこれはdomainの責務ってことで
ErrInvalidSortをdomainで落としてるって認識か。
cursor.goをusecaseに置いてるけどこれは別にここでいいの？結局具体がどこにあるかで責務はきまる？
逆にcursor.goはがっつりhttp含んでるからdomainで定義すると終わる
*/
/*domain Err だけで十分なら、usecase Err を作らないのが正解。usecaseのエラーは操作の失敗
認可：ErrPermissionDenied
リソース不存在：ErrNotFound
ドメインを横断する場合
*/
//入力不正系は wrap しない
