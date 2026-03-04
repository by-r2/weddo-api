CREATE TABLE IF NOT EXISTS guests (
    id             TEXT PRIMARY KEY,
    invitation_id  TEXT NOT NULL REFERENCES invitations(id) ON DELETE CASCADE,
    wedding_id     TEXT NOT NULL REFERENCES weddings(id),
    name           TEXT NOT NULL COLLATE NOCASE,
    phone          TEXT DEFAULT '',
    email          TEXT DEFAULT '',
    status         TEXT NOT NULL DEFAULT 'pending',
    confirmed_at   DATETIME,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_guests_invitation_id ON guests(invitation_id);
CREATE INDEX IF NOT EXISTS idx_guests_wedding_id ON guests(wedding_id);
CREATE INDEX IF NOT EXISTS idx_guests_wedding_status ON guests(wedding_id, status);
CREATE INDEX IF NOT EXISTS idx_guests_wedding_name ON guests(wedding_id, name COLLATE NOCASE);
