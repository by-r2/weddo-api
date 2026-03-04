CREATE TABLE IF NOT EXISTS google_integrations (
    wedding_id               TEXT PRIMARY KEY REFERENCES weddings(id) ON DELETE CASCADE,
    spreadsheet_id           TEXT NOT NULL,
    spreadsheet_url          TEXT NOT NULL,
    encrypted_access_token   TEXT NOT NULL,
    encrypted_refresh_token  TEXT NOT NULL,
    token_expiry             TIMESTAMPTZ,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_google_integrations_spreadsheet_id ON google_integrations(spreadsheet_id);
