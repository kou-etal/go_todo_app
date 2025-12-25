package taskrepo

import "time"

type TaskRecord struct {
	ID          string    `db:"id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Status      string    `db:"status"`
	DueDate     time.Time `db:"due_date"`
	Created     time.Time `db:"created_at"`
	Updated     time.Time `db:"updated_at"`
	Version     uint64    `db:"version"`
} //予約語危険、snake_caseが丸い。
//GoのstructはpascalCase
//DBはsnake_case
//jsonはcamelcase
