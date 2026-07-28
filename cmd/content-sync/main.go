// cmd/content-sync copies every project and post from a source Postgres
// database (read directly, the same way cmd/migrate does) to a destination
// blog instance's admin HTTP routes (POST /projects, POST /posts).
//
// This exists because a fresh admin-mode deployment's database is often not
// reachable from outside its own network (e.g. a k3s-internal Postgres,
// reachable only via the blog instance itself over Tailscale) — the running
// instance's own admin forms are the only door in, so this drives them the
// same way a human filling in the compose form would, rather than talking to
// the destination database at all.
//
// Idempotent: re-running against a destination that already has some or all
// of the content skips anything already present, detected via the same
// "already exists" duplicate-slug form error postCreatePost/projectCreatePost
// report to a human, rather than failing the whole run.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jonathanschwarzhaupt/home-blog/internal/database"
)

func main() {
	sourceDSN := flag.String("source-db-dsn", os.Getenv("BLOG_DB_DSN"), "PostgreSQL DSN to read projects/posts from")
	destURL := flag.String("dest-url", "", "Base URL of the destination blog instance, admin mode enabled (e.g. https://blog.example.ts.net)")
	dryRun := flag.Bool("dry-run", false, "List what would be created without making any destination requests")
	flag.Parse()

	if *sourceDSN == "" {
		fmt.Fprintln(os.Stderr, "source-db-dsn is required (set BLOG_DB_DSN or pass -source-db-dsn)")
		os.Exit(1)
	}
	if *destURL == "" {
		fmt.Fprintln(os.Stderr, "-dest-url is required, e.g. -dest-url=https://blog.example.ts.net")
		os.Exit(1)
	}

	if err := run(context.Background(), *sourceDSN, strings.TrimRight(*destURL, "/"), *dryRun); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, sourceDSN, destURL string, dryRun bool) error {
	pool, err := pgxpool.New(ctx, sourceDSN)
	if err != nil {
		return fmt.Errorf("connect to source db: %w", err)
	}
	defer pool.Close()
	q := database.New(pool)

	projects, err := q.ListProjects(ctx)
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	posts, err := q.ListPosts(ctx)
	if err != nil {
		return fmt.Errorf("list posts: %w", err)
	}

	if dryRun {
		for _, p := range projects {
			fmt.Printf("[dry-run] would create project %q\n", p.Name)
		}
		for _, p := range posts {
			fmt.Printf("[dry-run] would create post %q\n", p.Title)
		}
		return nil
	}

	client := &http.Client{}
	client.Jar, err = cookiejar.New(nil)
	if err != nil {
		return err
	}

	for _, p := range projects {
		form := url.Values{
			"name":        {p.Name},
			"description": {p.Description},
			"created_at":  {p.CreatedAt.Time.Format("2006-01-02")},
		}
		status, body, err := postForm(client, destURL+"/projects", form)
		if err != nil {
			return fmt.Errorf("create project %q: %w", p.Name, err)
		}
		switch classify(status, body) {
		case outcomeCreated:
			fmt.Printf("created project %q\n", p.Name)
		case outcomeSkipped:
			fmt.Printf("skipped project %q (already exists)\n", p.Name)
		case outcomeFailed:
			return fmt.Errorf("create project %q failed: HTTP %d\n%s", p.Name, status, snippet(body))
		}
	}

	projectIDs, err := discoverDestProjectIDs(client, destURL)
	if err != nil {
		return fmt.Errorf("discover destination project IDs: %w", err)
	}

	for _, p := range posts {
		linked, err := q.GetProjectsForPost(ctx, p.ID)
		if err != nil {
			return fmt.Errorf("get projects for post %q: %w", p.Title, err)
		}

		form := url.Values{
			"title":        {p.Title},
			"body":         {p.Body},
			"so_what":      {p.SoWhat},
			"tags":         {strings.Join(p.Tags, ", ")},
			"published_at": {p.PublishedAt.Time.Format("2006-01-02")},
		}
		for _, lp := range linked {
			if id, ok := projectIDs[lp.Name]; ok {
				form.Add("project_ids", id)
			}
		}

		status, body, err := postForm(client, destURL+"/posts", form)
		if err != nil {
			return fmt.Errorf("create post %q: %w", p.Title, err)
		}
		switch classify(status, body) {
		case outcomeCreated:
			fmt.Printf("created post %q\n", p.Title)
		case outcomeSkipped:
			fmt.Printf("skipped post %q (already exists)\n", p.Title)
		case outcomeFailed:
			return fmt.Errorf("create post %q failed: HTTP %d\n%s", p.Title, status, snippet(body))
		}
	}

	return nil
}

type outcome int

const (
	outcomeCreated outcome = iota
	outcomeSkipped
	outcomeFailed
)

// classify reads postCreatePost/projectCreatePost's actual response shapes:
// a successful create redirects through to a 200 (the client follows it), a
// rejected duplicate slug re-renders the form as 422 with a fixed
// "already exists" phrase in the body, and anything else is unexpected.
func classify(status int, body string) outcome {
	switch {
	case status == http.StatusOK:
		return outcomeCreated
	case status == http.StatusUnprocessableEntity && strings.Contains(body, "already exists"):
		return outcomeSkipped
	default:
		return outcomeFailed
	}
}

var (
	projectCheckboxRe = regexp.MustCompile(`(?s)<input name="project_ids" value="(\d+)"[^>]*>(.*?)</label>`)
	htmlTagRe         = regexp.MustCompile(`<[^>]*>`)
)

// discoverDestProjectIDs maps each project's name to its destination-assigned
// ID by scraping GET /posts/new's project checkboxes — there's no JSON API
// to ask for this, only the same HTML form a human uses, so this reads the
// project_ids checkbox list the same way postCreatePost's own form renders
// it (see ui/templ/pages/admin/post_create.templ's projectCheckboxes).
func discoverDestProjectIDs(client *http.Client, destURL string) (map[string]string, error) {
	resp, err := client.Get(destURL + "/posts/new")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseProjectCheckboxes(string(b)), nil
}

func parseProjectCheckboxes(html string) map[string]string {
	ids := map[string]string{}
	for _, m := range projectCheckboxRe.FindAllStringSubmatch(html, -1) {
		name := strings.TrimSpace(htmlTagRe.ReplaceAllString(m[2], ""))
		ids[name] = m[1]
	}
	return ids
}

func postForm(client *http.Client, target string, form url.Values) (status int, body string, err error) {
	req, err := http.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}
	return resp.StatusCode, string(b), nil
}

func snippet(s string) string {
	const max = 2000
	if len(s) > max {
		return s[:max]
	}
	return s
}
