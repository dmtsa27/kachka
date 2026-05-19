-- +goose Up
-- +goose StatementBegin
ALTER TABLE challenge_bootstrap 
    ADD COLUMN days_per_week INTEGER NOT NULL DEFAULT 3,
    ADD COLUMN duration_days INTEGER NOT NULL DEFAULT 180;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE challenge_bootstrap 
    DROP COLUMN IF EXISTS days_per_week,
    DROP COLUMN IF EXISTS duration_days;
-- +goose StatementEnd
