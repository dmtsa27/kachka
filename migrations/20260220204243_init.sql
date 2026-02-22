-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    telegram_id bigint primary key,
    username text,
    days_trained integer default 0,
    is_active boolean default true,
    failed_at TIMESTAMP WITH TIME ZONE
);
CREATE TABLE workouts(
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(telegram_id),
    workout_date TIMESTAMP WITH TIME ZONE
);
CREATE TABLE sessions(
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(telegram_id),
    chat_id BIGINT NOT NULL,
    message_id BIGINT NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE default NOW(),
    last_video_at TIMESTAMP WITH TIME ZONE,
    session_date DATE DEFAULT CURRENT_DATE
);

CREATE TABLE challenges(
    id serial primary key,
    days_per_week integer default 3,
    challenge_duration integer, 
    is_active boolean default true
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS workouts;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS challenges;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
