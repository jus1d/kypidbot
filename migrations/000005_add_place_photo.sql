-- +goose Up
ALTER TABLE places ADD COLUMN photo_url TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE places DROP COLUMN photo_url;
