CREATE TABLE IF NOT EXISTS users (
    id              TEXT PRIMARY KEY,
    wedding_id      TEXT REFERENCES weddings(id) ON DELETE CASCADE,
    name            TEXT NOT NULL DEFAULT '',
    email           TEXT NOT NULL,
    avatar_url      TEXT NOT NULL DEFAULT '',
    google_id       TEXT,
    password_hash   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id) WHERE google_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_wedding_id ON users(wedding_id);

-- Migrar admins existentes de weddings para users
INSERT INTO users (id, wedding_id, name, email, password_hash, created_at, updated_at)
SELECT
    gen_random_uuid()::text,
    id,
    partner1_name,
    admin_email,
    admin_pass_hash,
    created_at,
    updated_at
FROM weddings
WHERE admin_email IS NOT NULL AND admin_email != '';

ALTER TABLE weddings DROP COLUMN IF EXISTS admin_email;
ALTER TABLE weddings DROP COLUMN IF EXISTS admin_pass_hash;
