package create

type Command struct {
	//id作成はdomainの責務
	Title       string
	Description string
	DueDate     int //nilは期限なし。ゆえにポインタ、いや今は期限必須で。
}

type Result struct {
	ID string
}

/*type Command struct {
	//id作成はdomainの責務
	Title       string
	Description string
	DueDate     dtask.DueOption //nilは期限なし。ゆえにポインタ、いや今は期限必須で。
}*/
//これにするとrowの型とdomainの型が混ざって良くない。さらにhandlerでint->dueoptionへの変換が必須
//これをなくすためにcommandはrowの型でusecaseで変換するようにした。良くないかも。
