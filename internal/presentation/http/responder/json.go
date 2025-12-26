package responder

//helperとかutilsとかがそもそも責務分かりにくくて良くない。

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		// logger.Error("write json response failed", err)
	}
}

//err := json.NewEncoder(w).Encode(v)で一回 　streamを重視、2回に分けてもデータを中間処理しない場合は無駄
// _, err := fmt.Fprintf(w, "%s", bodyBytes)で二回に分ける　データを中間処理できる、middlewareとかで多い

/*import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrResponse struct {
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}//ここでErrResponse定義はよくない。Responderは判断しない。ゆえにcontextも使わない。
//contextを意味もなく持たせることを避ける。

func RespondJSON(ctx context.Context, w http.ResponseWriter, body any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("encode response error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		rsp := ErrResponse{
			Message: http.StatusText(http.StatusInternalServerError),
		}
		if err := json.NewEncoder(w).Encode(rsp); err != nil {
			fmt.Printf("write error response error: %v", err)
		}
		return
	}

	w.WriteHeader(status)
	if _, err := fmt.Fprintf(w, "%s", bodyBytes); err != nil {
		fmt.Printf("write response error: %v", err)
	}
}*/
