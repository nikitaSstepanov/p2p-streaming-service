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
    path VARCHAR(255)
);

-- +goose Down

DROP TABLE users;

DROP TABLE movies;