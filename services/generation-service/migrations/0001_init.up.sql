CREATE TABLE IF NOT EXISTS generation_jobs (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL,
    status          TEXT NOT NULL,
    prompt          TEXT NOT NULL,
    title           TEXT NOT NULL,
    genre           TEXT,
    mood            TEXT,
    language        TEXT,
    lyrics          TEXT,
    track_id        UUID,
    failure_reason  TEXT,
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS generation_jobs_user_idx ON generation_jobs (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS lyrics_versions (
    id          UUID PRIMARY KEY,
    job_id      UUID NOT NULL REFERENCES generation_jobs(id) ON DELETE CASCADE,
    revision    INT NOT NULL,
    body        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    UNIQUE (job_id, revision)
);
