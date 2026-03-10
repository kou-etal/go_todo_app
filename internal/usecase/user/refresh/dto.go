package refresh

type Command struct {
	RefreshToken string
}

type Result struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}
