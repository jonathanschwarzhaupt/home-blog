package blog

import (
	"net/url"
	"strconv"
)

const PostsPerPage = 7

// PostFilters is deliberately lenient parsing: an invalid/missing page or
// sort value just falls back to its default rather than producing a
// validation error — this is a read-only browsing page, not a form
// submission, so there's no field to show an error against.
type PostFilters struct {
	Page int
	Sort string // "newest" (default) or "oldest"
	From string // "YYYY-MM-DD" or ""
	To   string // "YYYY-MM-DD" or ""
	Tag  string // single tag, "" = no filter
}

func ParsePostFilters(query url.Values) PostFilters {
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	sort := query.Get("sort")
	if sort != "oldest" {
		sort = "newest"
	}

	return PostFilters{
		Page: page,
		Sort: sort,
		From: query.Get("from"),
		To:   query.Get("to"),
		Tag:  query.Get("tag"),
	}
}

// baseQuery carries forward the currently active sort/date-range/tag filters
// (but not page — callers decide whether page belongs in a given link).
func (f PostFilters) baseQuery() url.Values {
	v := url.Values{}
	if f.Sort == "oldest" {
		v.Set("sort", "oldest")
	}
	if f.From != "" {
		v.Set("from", f.From)
	}
	if f.To != "" {
		v.Set("to", f.To)
	}
	if f.Tag != "" {
		v.Set("tag", f.Tag)
	}
	return v
}

func linkFrom(v url.Values) string {
	if len(v) == 0 {
		return "/posts"
	}
	return "/posts?" + v.Encode()
}

// TagFilterLink builds a link filtering /posts to a single tag, from
// anywhere a tag is shown standalone (a post card, a post's own page) rather
// than from the posts index itself — so it doesn't try to preserve unrelated
// filter state from wherever it was clicked, just a fresh, correctly
// URL-encoded /posts?tag=... link.
func TagFilterLink(tag string) string {
	v := url.Values{}
	v.Set("tag", tag)
	return linkFrom(v)
}

// SortLink builds a link that switches to the given sort direction,
// preserving the active date range and tag. Deliberately omits page —
// switching sort changes the order of every result, so "page N" from the
// old view doesn't mean anything in the new one.
func (f PostFilters) SortLink(sort string) string {
	v := url.Values{}
	if sort == "oldest" {
		v.Set("sort", "oldest")
	}
	if f.From != "" {
		v.Set("from", f.From)
	}
	if f.To != "" {
		v.Set("to", f.To)
	}
	if f.Tag != "" {
		v.Set("tag", f.Tag)
	}
	return linkFrom(v)
}

// TagLink builds a link that switches to the given tag, preserving sort and
// date range, resetting page for the same reason SortLink does.
func (f PostFilters) TagLink(tag string) string {
	v := url.Values{}
	if f.Sort == "oldest" {
		v.Set("sort", "oldest")
	}
	if f.From != "" {
		v.Set("from", f.From)
	}
	if f.To != "" {
		v.Set("to", f.To)
	}
	if tag != "" {
		v.Set("tag", tag)
	}
	return linkFrom(v)
}

// PageLink builds a link to a specific page, preserving sort, date range,
// and tag. Unlike SortLink/TagLink, plain pagination doesn't invalidate the
// current filters.
func (f PostFilters) PageLink(page int) string {
	v := f.baseQuery()
	if page > 1 {
		v.Set("page", strconv.Itoa(page))
	}
	return linkFrom(v)
}

// ClearDateRangeLink preserves sort and tag but drops from/to and resets to
// page 1.
func (f PostFilters) ClearDateRangeLink() string {
	v := url.Values{}
	if f.Sort == "oldest" {
		v.Set("sort", "oldest")
	}
	if f.Tag != "" {
		v.Set("tag", f.Tag)
	}
	return linkFrom(v)
}
