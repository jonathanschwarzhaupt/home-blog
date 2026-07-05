package main

import (
	"net/http"

	"github.com/a-h/templ"
)

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error(err.Error(), "request_id", requestIDFromContext(r.Context()), "method", r.Method, "uri", r.URL.RequestURI())
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// notFound is Let's Go's own helper shape (helpers.go, chapter 03.04) —
// clientError(w, 404) under a semantic name, so handlers signal "not found"
// the same way they already signal any other client error, rather than
// calling net/http's http.NotFound directly. It doesn't know anything about
// the styled page; styleNotFound (notfound.go) upgrades the resulting 404
// response, decoupled from whichever code path produced it.
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, component templ.Component) {
	w.WriteHeader(status)
	if err := component.Render(r.Context(), w); err != nil {
		app.logger.Error(err.Error(), "request_id", requestIDFromContext(r.Context()), "method", r.Method, "uri", r.URL.RequestURI())
	}
}
