-- Starter schema for the V1 catalog, accounts, and saved preferences.
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS users_username_nocase ON users(lower(username));
CREATE UNIQUE INDEX IF NOT EXISTS users_email_nocase ON users(lower(email));

CREATE TABLE IF NOT EXISTS sessions (
    token_hash TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS sessions_user_id ON sessions(user_id);

CREATE TABLE IF NOT EXISTS titles (
    id INTEGER PRIMARY KEY,
    provider_id TEXT UNIQUE,
    kind TEXT NOT NULL CHECK (kind IN ('movie', 'show')),
    name TEXT NOT NULL,
    release_year INTEGER,
    synopsis TEXT,
    poster_url TEXT,
    trailer_url TEXT
);

CREATE TABLE IF NOT EXISTS services (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS genres (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS title_genres (
    title_id INTEGER NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    genre_id INTEGER NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (title_id, genre_id)
);

CREATE TABLE IF NOT EXISTS title_services (
    title_id INTEGER NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    PRIMARY KEY (title_id, service_id)
);

CREATE TABLE IF NOT EXISTS user_titles (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title_id INTEGER NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, title_id)
);

CREATE TABLE IF NOT EXISTS user_services (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, service_id)
);

CREATE TABLE IF NOT EXISTS ratings (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title_id INTEGER NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    stars INTEGER NOT NULL CHECK (stars BETWEEN 1 AND 5),
    review TEXT NOT NULL DEFAULT '',
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, title_id)
);

CREATE INDEX IF NOT EXISTS ratings_title_id ON ratings(title_id);

CREATE TABLE IF NOT EXISTS user_watched (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title_id INTEGER NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    watched_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, title_id)
);

CREATE INDEX IF NOT EXISTS title_genres_genre_id ON title_genres(genre_id);
CREATE INDEX IF NOT EXISTS title_services_service_id ON title_services(service_id);
CREATE INDEX IF NOT EXISTS titles_release_year ON titles(release_year);

INSERT OR IGNORE INTO services (name) VALUES
    ('Netflix'),
    ('Hulu'),
    ('Amazon Prime Video'),
    ('HBO Max'),
    ('Disney Plus'),
    ('Paramount Plus Essential'),
    ('Paramount Plus Premium'),
    ('Peacock Premium'),
    ('Peacock Premium Plus'),
    ('Apple TV');
