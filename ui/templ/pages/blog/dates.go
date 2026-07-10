package blog

import "time"

// FormatPostDate renders a Post's PublishedAt as the ISO date shown next to
// its tags — matching the monospace treatment tags/nav/titles already use
// (ADR-0005) rather than a conventional "Jan 2, 2006" blog date.
func FormatPostDate(t time.Time) string {
	return t.Format("2006-01-02")
}
