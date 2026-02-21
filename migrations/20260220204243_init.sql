-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    telegram_id bigint primary key,
    username text,
    is_active boolean default true --потім можна поміняти на фолс коли буду емплементувати фічу де по реакції можна буде підтвердити участь в челенджі
);
CREATE TABLE workouts(
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(telegram_id),
    workout_date TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE challenges(
    id serial primary key,
    days_per_week integer default 3,
    challenge_duration integer, 
    is_active boolean default true --потім можна поміняти на фолс коли буду емплементувати фічу де по реакції можна буде підтвердити участь в челенджі
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS workouts;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS challenges;
-- +goose StatementEnd
