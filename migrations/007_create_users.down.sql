ALTER TABLE weddings ADD COLUMN IF NOT EXISTS admin_email TEXT NOT NULL DEFAULT '';
ALTER TABLE weddings ADD COLUMN IF NOT EXISTS admin_pass_hash TEXT NOT NULL DEFAULT '';

-- Restaurar admin_email/admin_pass_hash a partir do primeiro user de cada wedding
UPDATE weddings w
SET admin_email    = u.email,
    admin_pass_hash = COALESCE(u.password_hash, '')
FROM (
    SELECT DISTINCT ON (wedding_id) wedding_id, email, password_hash
    FROM users
    ORDER BY wedding_id, created_at ASC
) u
WHERE w.id = u.wedding_id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_weddings_admin_email ON weddings(admin_email);

DROP TABLE IF EXISTS users;
