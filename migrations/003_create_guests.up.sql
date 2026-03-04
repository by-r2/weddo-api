CREATE TABLE IF NOT EXISTS guests (
    id             TEXT PRIMARY KEY,
    invitation_id  TEXT NOT NULL REFERENCES invitations(id) ON DELETE CASCADE,
    wedding_id     TEXT NOT NULL REFERENCES weddings(id),
    name           TEXT NOT NULL,
    phone          TEXT DEFAULT '',
    email          TEXT DEFAULT '',
    status         TEXT NOT NULL DEFAULT 'pending',
    confirmed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_guests_invitation_id ON guests(invitation_id);
CREATE INDEX IF NOT EXISTS idx_guests_wedding_id ON guests(wedding_id);
CREATE INDEX IF NOT EXISTS idx_guests_wedding_status ON guests(wedding_id, status);
CREATE INDEX IF NOT EXISTS idx_guests_wedding_name ON guests(wedding_id, LOWER(name));
