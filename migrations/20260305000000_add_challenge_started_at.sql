-- +goose Up
-- +goose StatementBegin
ALTER TABLE challenges ADD COLUMN started_at TIMESTAMP WITH TIME ZONE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE challenges DROP COLUMN started_at;
-- +goose StatementEnd
