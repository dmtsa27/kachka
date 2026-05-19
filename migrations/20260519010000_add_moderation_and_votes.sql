-- +goose Up
-- +goose StatementBegin
ALTER TABLE workouts 
    ADD COLUMN is_cancelled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN cancelled_by BIGINT,
    ADD COLUMN cancelled_at TIMESTAMP WITH TIME ZONE;

CREATE TABLE votes (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    target_user_id BIGINT NOT NULL,
    initiator_id BIGINT NOT NULL,
    poll_id TEXT NOT NULL UNIQUE,
    amount INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    is_success BOOLEAN NOT NULL DEFAULT FALSE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS votes;

ALTER TABLE workouts 
    DROP COLUMN IF EXISTS is_cancelled,
    DROP COLUMN IF EXISTS cancelled_by,
    DROP COLUMN IF EXISTS cancelled_at;
-- +goose StatementEnd
