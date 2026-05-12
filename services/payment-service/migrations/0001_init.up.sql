CREATE TABLE IF NOT EXISTS payments (
    id                  UUID PRIMARY KEY,
    user_id             UUID NOT NULL,
    provider            TEXT NOT NULL,
    provider_payment_id TEXT,
    status              TEXT NOT NULL,
    price_currency      TEXT NOT NULL,
    price_amount        TEXT NOT NULL,
    pay_currency        TEXT,
    pay_address         TEXT,
    pay_amount          TEXT,
    purpose             TEXT NOT NULL,
    invoice_url         TEXT,
    created_at          TIMESTAMPTZ NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS payments_user_idx ON payments (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS payments_provider_idx ON payments (provider_payment_id);

CREATE TABLE IF NOT EXISTS payments_events (
    id          BIGSERIAL PRIMARY KEY,
    payment_id  UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    status      TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    raw_payload JSONB NOT NULL,
    UNIQUE (payment_id, status, occurred_at)
);
