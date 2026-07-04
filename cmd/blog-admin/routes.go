package main

import (
	"net/http"

	"github.com/justinas/alice"

	"github.com/jonathanschwarzhaupt/my-blog/ui"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.FileServerFS(ui.Files))
	mux.HandleFunc("GET /health", app.healthcheck)

	dynamic := alice.New(preventCSRF, app.sessionManager.LoadAndSave)

	mux.Handle("GET /posts/new", dynamic.ThenFunc(app.postCreate))
	mux.Handle("POST /posts", dynamic.ThenFunc(app.postCreatePost))

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	return standard.Then(mux)
}
