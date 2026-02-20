package list

import "time"

type Query struct {
	Limit int

	Sort   string
	Cursor string
}

type Result struct {
	Items      []Item
	NextCursor string
}

type Item struct {
	ID          string
	Title       string
	Description string
	Status      string

	DueDate *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	Version   int64
}
