package router

import (
	"net/http"
)

type Deps struct {
	Task   TaskDeps
	User   UserDeps
	AuthMW func(http.Handler) http.Handler
}

type TaskDeps struct {
	List   http.Handler
	Create http.Handler
	Update http.Handler
	Delete http.Handler
}
type UserDeps struct {
	Register http.Handler
	Login    http.Handler
	Refresh  http.Handler
}

func New(deps Deps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/health", http.HandlerFunc(healthHandler))

	mux.Handle("/tasks", deps.AuthMW(tasksHandler(deps.Task)))

	mux.Handle("/tasks/", deps.AuthMW(taskHandler(deps.Task)))

	mux.Handle("/users", usersHandler(deps.User))
	mux.Handle("/users/login", loginHandler(deps.User))
	mux.Handle("/users/refresh", refreshHandler(deps.User))

	return mux
}
