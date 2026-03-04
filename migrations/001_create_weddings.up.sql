CREATE TABLE IF NOT EXISTS weddings (
    id              TEXT PRIMARY KEY,
    slug            TEXT NOT NULL,
    title           TEXT NOT NULL,
    date            DATE,
    partner1_name   TEXT NOT NULL,
    partner2_name   TEXT NOT NULL,
    admin_email     TEXT NOT NULL,
    admin_pass_hash TEXT NOT NULL,
    active          INTEGER NOT NULL DEFAULT 1,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_weddings_slug ON weddings(slug);
CREATE UNIQUE INDEX IF NOT EXISTS idx_weddings_admin_email ON weddings(admin_email);
