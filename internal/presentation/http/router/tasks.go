package router

import "net/http"

//共有しないから大文字ではない
//t.List.ServeHTTP(w, r)を返す
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
		//idの抽出はhandlerに寄せてる
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

//REST的にはtasksCollection、tasksItemでもいいがこれはわかりにくい。
