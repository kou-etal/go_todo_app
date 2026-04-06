package login

type Command struct {
	Email    string
	Password string
}

type Result struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}
