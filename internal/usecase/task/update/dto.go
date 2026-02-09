package update

type Command struct { //commandにはhandlerからraw dataが送られてくる->最低限弾いたデータ
	ID      string //これ識別とかじゃなくて更新する場所のID
	Version uint64
	//updateはcreateと違ってoptional。ゆえにポインタ
	Title       *string
	Description *string
	DueDate     *int //DueOption(7/14/21/30)
}

type Result struct {
	ID string
} //TODO:これversion返す
