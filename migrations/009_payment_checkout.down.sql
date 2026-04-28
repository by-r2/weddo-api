DROP TABLE IF EXISTS payment_items CASCADE;

ALTER TABLE gifts DROP CONSTRAINT IF EXISTS gifts_kind_check;
DROP INDEX IF EXISTS idx_gifts_one_cash_per_wedding;

DELETE FROM gifts WHERE kind = 'cash_template';

ALTER TABLE gifts DROP COLUMN IF EXISTS kind;

ALTER TABLE payments ADD COLUMN IF NOT EXISTS gift_id TEXT REFERENCES gifts (id);
