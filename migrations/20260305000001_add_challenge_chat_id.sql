-- +goose Up
-- +goose StatementBegin
ALTER TABLE challenges ADD COLUMN chat_id BIGINT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE challenges DROP COLUMN chat_id;
-- +goose StatementEnd
