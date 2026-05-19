-- +goose Up
-- +goose StatementBegin
ALTER TABLE challenges ADD COLUMN last_weekly_check_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE challenges ADD COLUMN last_daily_stats_at TIMESTAMP WITH TIME ZONE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE challenges DROP COLUMN last_weekly_check_at;
ALTER TABLE challenges DROP COLUMN last_daily_stats_at;
-- +goose StatementEnd
