-- +goose Up

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255),
    password VARCHAR(255),
    role VARCHAR(255)
);

CREATE TABLE movies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    paths VARCHAR(255),
    fileVersion INTEGER
);

CREATE TABLE adapters (
    id SERIAL PRIMARY KEY,
    movieId BIGINT,
    version INTEGER,
    length BIGINT,
    pieceLength BIGINT
);

INSERT INTO users (username, password, role) VALUES ('admin', '$2a$10$uO5L5aVpKAnteAwJgA3e0eo.pOdGclPLodcB8yKAkIEELTAeIz/ii', 'SUPER_ADMIN');

-- +goose Down

DROP TABLE users;

DROP TABLE movies;

DROP TABLE adapters;