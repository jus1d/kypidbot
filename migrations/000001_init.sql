-- +goose Up
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
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
    is_admin BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE pairs (
    id BIGSERIAL PRIMARY KEY,
    dill_id BIGINT NOT NULL REFERENCES users(id),
    doe_id BIGINT NOT NULL REFERENCES users(id),
    score DOUBLE PRECISION NOT NULL,
    time_intersection TEXT NOT NULL,
    is_fullmatch BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE places (
    id BIGSERIAL PRIMARY KEY,
    description TEXT NOT NULL
);

CREATE TABLE meetings (
    id BIGSERIAL PRIMARY KEY,
    pair_id BIGINT NOT NULL REFERENCES pairs(id),
    place_id BIGINT NOT NULL REFERENCES places(id),
    time TEXT NOT NULL,
    dill_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    doe_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    dill_cancelled BOOLEAN NOT NULL DEFAULT FALSE,
    doe_cancelled BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose Down
DROP TABLE IF EXISTS meetings;
DROP TABLE IF EXISTS pairs;
DROP TABLE IF EXISTS places;
DROP TABLE IF EXISTS users;
