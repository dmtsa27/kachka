-- +goose Up
-- +goose StatementBegin
ALTER TABLE votes ADD COLUMN type TEXT NOT NULL DEFAULT 'subtract';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE votes DROP COLUMN IF EXISTS type;
-- +goose StatementEnd
