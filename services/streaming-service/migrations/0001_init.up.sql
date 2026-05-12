CREATE TABLE IF NOT EXISTS track_audio (
    track_id     UUID PRIMARY KEY,
    object_key   TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes   BIGINT NOT NULL,
    uploaded_by  UUID,
    uploaded_at  TIMESTAMPTZ NOT NULL
);
