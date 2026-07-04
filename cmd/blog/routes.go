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
	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /posts/{slug}", app.postView)

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	return standard.Then(mux)
}
