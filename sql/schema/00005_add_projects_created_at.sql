-- +goose Up
ALTER TABLE projects ADD COLUMN created_at timestamptz NOT NULL DEFAULT now();

CREATE INDEX projects_created_at_idx ON projects (created_at DESC, id ASC);

-- +goose Down
DROP INDEX IF EXISTS projects_created_at_idx;
ALTER TABLE projects DROP COLUMN IF EXISTS created_at;
