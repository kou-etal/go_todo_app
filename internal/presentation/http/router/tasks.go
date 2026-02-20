package router

import "net/http"

func tasksHandler(t TaskDeps) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			t.List.ServeHTTP(w, r)
		case http.MethodPost:
			t.Create.ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func taskHandler(t TaskDeps) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodPatch:
			t.Update.ServeHTTP(w, r)
		case http.MethodDelete:
			t.Delete.ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
