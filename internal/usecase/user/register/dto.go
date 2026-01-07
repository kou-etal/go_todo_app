package register

type Command struct {
	Email    string
	Password string
	UserName string
}

type Result struct {
	UserID string
} //何返すべきかは議論。tokenのplainはここで返さずにmailerに含める。
