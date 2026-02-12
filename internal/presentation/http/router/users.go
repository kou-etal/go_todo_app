package router

import "net/http"

func usersHandler(u UserDeps) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			u.Register.ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
