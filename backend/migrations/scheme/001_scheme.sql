-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE,
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
    movieId SERIAL,
    version INTEGER,
    length INTEGER,
    pieceLength INTEGER,
    FOREIGN KEY (movieId) REFERENCES movies (id) ON DELETE CASCADE
);

CREATE TABLE playlists (
    id SERIAL PRIMARY KEY,
    userId SERIAL,
    title VARCHAR(255),
    FOREIGN KEY (userId) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    movieId SERIAL,
    userId SERIAL,
    text VARCHAR(255),
    FOREIGN KEY (movieId) REFERENCES movies (id) ON DELETE CASCADE,
    FOREIGN KEY (userId) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE movies_playlists (
    playlistId SERIAL,
    movieId SERIAL,
    FOREIGN KEY (playlistId) REFERENCES playlists (id) ON DELETE CASCADE,
    FOREIGN KEY (movieId) REFERENCES movies (id) ON DELETE CASCADE
);

CREATE TABLE tokens (
    id SERIAL PRIMARY KEY,
    userId SERIAL,
    token VARCHAR(255) UNIQUE,
    FOREIGN KEY (userId) REFERENCES users (id) ON DELETE CASCADE
);

INSERT INTO users (username, password, role) VALUES ('admin', '$2a$10$uO5L5aVpKAnteAwJgA3e0eo.pOdGclPLodcB8yKAkIEELTAeIz/ii', 'SUPER_ADMIN');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;

DROP TABLE movies;

DROP TABLE adapters;

DROP TABLE playlists;

DROP TABLE movies_playlists;

DROP TABLE comments;

DROP TABLE tokens;
-- +goose StatementEnd