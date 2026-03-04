CREATE TABLE IF NOT EXISTS invitations (
    id          TEXT PRIMARY KEY,
    wedding_id  TEXT NOT NULL REFERENCES weddings(id),
    code        TEXT NOT NULL,
    label       TEXT NOT NULL,
    max_guests  INTEGER NOT NULL DEFAULT 1,
    notes       TEXT DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_invitations_wedding_id ON invitations(wedding_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_invitations_wedding_code ON invitations(wedding_id, code);
