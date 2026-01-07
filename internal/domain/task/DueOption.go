package task

//TODO:これはUI都合のルール、ゆえにここで宣言不可。usecaseで宣言し、parseをそこで作ってfactoryのnewにはintを与える
type DueOption int

const (
	Due7Days  DueOption = 7
	Due14Days DueOption = 14
	Due21Days DueOption = 21
	Due30Days DueOption = 30
)
