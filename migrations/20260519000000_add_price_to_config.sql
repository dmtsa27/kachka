-- +goose Up
-- +goose StatementBegin
ALTER TABLE challenge_bootstrap ADD COLUMN price INTEGER NOT NULL DEFAULT 500;
ALTER TABLE challenges ADD COLUMN price INTEGER NOT NULL DEFAULT 500;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE challenge_bootstrap DROP COLUMN IF EXISTS price;
ALTER TABLE challenges DROP COLUMN IF EXISTS price;
-- +goose StatementEnd
