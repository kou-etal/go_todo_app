package list

//これはいいパッケージ名

import "time"

type Query struct {
	Limit int
	//limitをユーザーに決めさせるの良くない
	Sort   string
	Cursor string
	//""は初回取得を表す
}

type Result struct {
	Items      []Item
	NextCursor string
}

//entity->DTO
type Item struct {
	ID          string
	Title       string
	Description string
	Status      string

	DueDate *time.Time
	//nullがあり得るからポインタ
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   int64
} //これはただのデータ運搬やからカプセルにしない。
