package router

import (
	"net/http"
)

type Deps struct {
	Task TaskDeps
	User UserDeps
}

type TaskDeps struct {
	List   http.Handler
	Create http.Handler
	Update http.Handler
	Delete http.Handler
}
type UserDeps struct {
	Register http.Handler
}

func New(deps Deps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/health", http.HandlerFunc(healthHandler))

	mux.Handle("/tasks", tasksHandler(deps.Task))

	mux.Handle("/tasks/", taskHandler(deps.Task))

	mux.Handle("/users", usersHandler(deps.User))

	return mux
}
