CREATE EXTENSION IF NOT EXISTS "pg_trgm";

CREATE TABLE IF NOT EXISTS artists (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    country     TEXT NOT NULL DEFAULT '',
    biography   TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_artists_name ON artists USING gin (name gin_trgm_ops);

CREATE TABLE IF NOT EXISTS albums (
    id          UUID PRIMARY KEY,
    artist_id   UUID NOT NULL REFERENCES artists(id),
    title       TEXT NOT NULL,
    year        INT NOT NULL DEFAULT 0,
    cover_url   TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_albums_artist_id ON albums (artist_id);
CREATE INDEX IF NOT EXISTS idx_albums_title ON albums USING gin (title gin_trgm_ops);

CREATE TABLE IF NOT EXISTS tracks (
    id              UUID PRIMARY KEY,
    album_id        UUID NOT NULL REFERENCES albums(id),
    artist_id       UUID NOT NULL REFERENCES artists(id),
    title           TEXT NOT NULL,
    duration_sec    INT NOT NULL DEFAULT 0,
    isrc            TEXT UNIQUE,
    genres          TEXT[] NOT NULL DEFAULT '{}',
    user_generated  BOOLEAN NOT NULL DEFAULT FALSE,
    owner_user_id   UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tracks_album_id ON tracks (album_id);
CREATE INDEX IF NOT EXISTS idx_tracks_artist_id ON tracks (artist_id);
CREATE INDEX IF NOT EXISTS idx_tracks_title ON tracks USING gin (title gin_trgm_ops);
