package register

type Command struct {
	Email    string
	Password string
	UserName string
}

type Result struct {
	UserID string
}
