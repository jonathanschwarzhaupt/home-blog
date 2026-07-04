-- +goose Up
CREATE TABLE posts (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title text NOT NULL,
    slug text NOT NULL UNIQUE,
    body text NOT NULL,
    so_what text NOT NULL,
    tags text[] NOT NULL DEFAULT '{}',
    version integer NOT NULL DEFAULT 1,
    published_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT posts_so_what_not_blank CHECK (so_what <> '')
);

CREATE INDEX posts_published_at_idx ON posts (published_at DESC, id ASC);

-- +goose Down
DROP TABLE IF EXISTS posts;
