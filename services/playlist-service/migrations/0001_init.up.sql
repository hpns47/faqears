CREATE TABLE IF NOT EXISTS playlists (
    id          UUID PRIMARY KEY,
    owner_id    UUID NOT NULL,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    public      BOOLEAN NOT NULL DEFAULT FALSE,
    permalink   TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS playlists_owner_idx ON playlists (owner_id);

CREATE TABLE IF NOT EXISTS playlist_tracks (
    playlist_id UUID NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    track_id    UUID NOT NULL,
    rank        NUMERIC NOT NULL,
    added_by    UUID NOT NULL,
    added_at    TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (playlist_id, track_id)
);
CREATE INDEX IF NOT EXISTS playlist_tracks_rank_idx ON playlist_tracks (playlist_id, rank);

CREATE TABLE IF NOT EXISTS playlist_collaborators (
    playlist_id UUID NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL,
    added_at    TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (playlist_id, user_id)
);
