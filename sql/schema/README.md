# Schema

Versioned goose migrations, sequentially numbered (`NNNNN_name.sql`), each containing `-- +goose Up` / `-- +goose Down` sections. Embedded into `cmd/migrate` at build time (see `embed.go`) so migrations ship inside that binary rather than being read from disk at runtime — this is what runs as the Kubernetes init container ahead of `blog`/`blog-admin`.

Run with `make db/migrations/up` (or `db/migrations/down`, `db/migrations/status`).
