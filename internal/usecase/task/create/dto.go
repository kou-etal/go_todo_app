package create

type Command struct {
	UserID      string
	Title       string
	Description string
	DueDate     int
}

type Result struct {
	ID string
}
