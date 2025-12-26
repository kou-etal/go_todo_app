package responder

type ErrResponse struct {
	Message string `json:"message"`
} //これ大文字にしないと別パッケージから使えない。
