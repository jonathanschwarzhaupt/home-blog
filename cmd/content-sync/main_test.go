package main

import (
	"net/http"
	"testing"

	"github.com/jonathanschwarzhaupt/home-blog/internal/assert"
)

func TestParseProjectCheckboxes(t *testing.T) {
	// A trimmed real fixture from GET /posts/new's rendered checkbox
	// fieldset (see ui/templ/pages/admin/post_create.templ's
	// projectCheckboxes) — nested icon <div>/<svg> markup sits between the
	// <input> and the visible label text, which is what the parser has to
	// see through.
	html := `<fieldset class="flex flex-col gap-2"><legend class="text-sm font-medium">Projects</legend>` +
		`<label class="flex items-center gap-2 text-sm"><div class="relative inline-flex items-center">` +
		`<input name="project_ids" value="1" type="checkbox" class="peer"></div>Home Plumbing</label>` +
		`<label class="flex items-center gap-2 text-sm"><div class="relative inline-flex items-center">` +
		`<input name="project_ids" value="2" type="checkbox" class="peer"></div>Homelab</label></fieldset>`

	got := parseProjectCheckboxes(html)

	assert.Equal(t, got["Home Plumbing"], "1")
	assert.Equal(t, got["Homelab"], "2")
	assert.Equal(t, len(got), 2)
}

func TestParseProjectCheckboxes_NoProjects(t *testing.T) {
	// projectCheckboxes renders nothing at all when allProjects is empty.
	got := parseProjectCheckboxes(`<form method="POST" action="/posts"></form>`)

	assert.Equal(t, len(got), 0)
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   outcome
	}{
		{"success", http.StatusOK, "Post published", outcomeCreated},
		{"duplicate post", http.StatusUnprocessableEntity, "A post with this title (or a very similar one) already exists — try a different title.", outcomeSkipped},
		{"duplicate project", http.StatusUnprocessableEntity, "A project with this name (or a very similar one) already exists — try a different name.", outcomeSkipped},
		{"other validation error", http.StatusUnprocessableEntity, "This field cannot be blank", outcomeFailed},
		{"server error", http.StatusInternalServerError, "Internal Server Error", outcomeFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, classify(tt.status, tt.body), tt.want)
		})
	}
}
