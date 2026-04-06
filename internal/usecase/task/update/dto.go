package update

type Command struct {
	UserID  string
	ID      string
	Version uint64

	Title       *string
	Description *string
	DueDate     *int
}

type Result struct {
	ID string
} //TODO:これversion返す
