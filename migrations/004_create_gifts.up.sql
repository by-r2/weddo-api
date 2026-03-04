CREATE TABLE IF NOT EXISTS gifts (
    id          TEXT PRIMARY KEY,
    wedding_id  TEXT NOT NULL REFERENCES weddings(id),
    name        TEXT NOT NULL,
    description TEXT DEFAULT '',
    price       DOUBLE PRECISION NOT NULL,
    image_url   TEXT DEFAULT '',
    category    TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'available',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_gifts_wedding_id ON gifts(wedding_id);
CREATE INDEX IF NOT EXISTS idx_gifts_category ON gifts(wedding_id, category);
CREATE INDEX IF NOT EXISTS idx_gifts_status ON gifts(wedding_id, status);
