CREATE TABLE IF NOT EXISTS play_events (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID NOT NULL,
    track_id    UUID NOT NULL,
    session_id  UUID,
    event_type  TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS play_events_user_idx ON play_events (user_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS play_events_track_idx ON play_events (track_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS play_events_user_track_idx ON play_events (user_id, track_id);
