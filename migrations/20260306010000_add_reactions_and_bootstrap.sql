-- +goose Up
-- +goose StatementBegin
ALTER TABLE workouts
    ADD COLUMN chat_id BIGINT,
    ADD COLUMN completion_message_id BIGINT;

CREATE TABLE challenge_bootstrap (
    chat_id BIGINT PRIMARY KEY,
    welcome_message_id BIGINT NOT NULL,
    roster_frozen_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    is_started BOOLEAN NOT NULL DEFAULT FALSE,
    started_at TIMESTAMP WITH TIME ZONE,
    is_bot_admin BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE chat_members (
    chat_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    is_bot BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    first_seen_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (chat_id, user_id)
);

CREATE TABLE message_reactions (
    chat_id BIGINT NOT NULL,
    message_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    emoji TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (chat_id, message_id, user_id, emoji)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS message_reactions;
DROP TABLE IF EXISTS chat_members;
DROP TABLE IF EXISTS challenge_bootstrap;

ALTER TABLE workouts
    DROP COLUMN IF EXISTS completion_message_id,
    DROP COLUMN IF EXISTS chat_id;
-- +goose StatementEnd
