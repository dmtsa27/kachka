-- +goose Up
-- +goose StatementBegin
ALTER TABLE challenge_bootstrap ADD COLUMN expected_reactions INTEGER NOT NULL DEFAULT 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE challenge_bootstrap DROP COLUMN expected_reactions;
-- +goose StatementEnd
