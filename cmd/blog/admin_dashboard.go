package main

import (
	"net/http"

	"github.com/jonathanschwarzhaupt/my-blog/ui/templ/pages/admin"
)

func (app *application) adminDashboard(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, admin.Dashboard())
}
