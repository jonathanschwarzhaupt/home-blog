package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/jonathanschwarzhaupt/my-blog/internal/assert"
	"github.com/jonathanschwarzhaupt/my-blog/internal/database"
	"github.com/jonathanschwarzhaupt/my-blog/internal/database/mocks"
)

func TestHome_ListsPostsNewestFirst(t *testing.T) {
	newer := pgtype.Timestamptz{Time: time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC), Valid: true}
	older := pgtype.Timestamptz{Time: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), Valid: true}

	mockDB := &mocks.MockQuerier{
		ListPostsFunc: func(ctx context.Context) ([]database.Post, error) {
			return []database.Post{
				{ID: 2, Title: "Newer Post", Slug: "newer-post", SoWhat: "recent", PublishedAt: newer},
				{ID: 1, Title: "Older Post", Slug: "older-post", SoWhat: "past", PublishedAt: older},
			}, nil
		},
	}

	app := newTestApplicationWithDB(mockDB)

	ts := httptest.NewServer(app.routes())
	defer ts.Close()

	rs, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, rs.StatusCode, http.StatusOK)

	html := string(body)
	newerIdx := strings.Index(html, "Newer Post")
	olderIdx := strings.Index(html, "Older Post")

	assert.True(t, newerIdx >= 0)
	assert.True(t, olderIdx >= 0)
	assert.True(t, newerIdx < olderIdx)
}
