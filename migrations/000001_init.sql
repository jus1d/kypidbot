-- +goose Up
CREATE TYPE confirmation_state AS ENUM ('not_confirmed', 'confirmed', 'cancelled');

CREATE TABLE users (
    telegram_id BIGINT PRIMARY KEY,
    username TEXT,
    first_name TEXT,
    last_name TEXT,
    is_bot BOOLEAN NOT NULL DEFAULT FALSE,
    language_code TEXT,
    is_premium BOOLEAN NOT NULL DEFAULT FALSE,
    sex TEXT,
    about TEXT NOT NULL DEFAULT '',
    state TEXT NOT NULL DEFAULT 'start',
    time_ranges TEXT NOT NULL DEFAULT '000000',
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    referral_code TEXT UNIQUE,
    referrer_id BIGINT REFERENCES users(telegram_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE places (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL
);

CREATE TABLE meetings (
    id SERIAL PRIMARY KEY,
    dill_id BIGINT NOT NULL REFERENCES users(telegram_id),
    doe_id BIGINT NOT NULL REFERENCES users(telegram_id),
    pair_score DOUBLE PRECISION NOT NULL,
    is_fullmatch BOOLEAN NOT NULL DEFAULT FALSE,
    place_id INTEGER REFERENCES places(id),
    time TEXT,
    dill_state confirmation_state NOT NULL DEFAULT 'not_confirmed',
    doe_state confirmation_state NOT NULL DEFAULT 'not_confirmed'
);

-- +goose Down
DROP TABLE IF EXISTS meetings;
DROP TABLE IF EXISTS places;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS confirmation_state;
