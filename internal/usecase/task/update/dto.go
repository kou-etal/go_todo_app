package update

type Command struct {
	ID      string
	Version uint64
	//updateはcreateと違ってoptional。ゆえにポインタ
	Title       *string
	Description *string
	DueDate     *int //DueOption(7/14/21/30)
}

type Result struct {
	ID string
}
