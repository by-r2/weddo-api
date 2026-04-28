-- Modelo checkout: tipo de gift, linhas por pagamento, contribuição em dinheiro como gift template único por casamento.

-- 1. gifts.kind (catálogo vs template de contribuição em dinheiro)
ALTER TABLE gifts ADD COLUMN kind TEXT NOT NULL DEFAULT 'catalog';
ALTER TABLE gifts ADD CONSTRAINT gifts_kind_check CHECK (kind IN ('catalog', 'cash_template'));

CREATE UNIQUE INDEX idx_gifts_one_cash_per_wedding ON gifts (wedding_id) WHERE kind = 'cash_template';

-- 2. Um cash_template por wedding (ID determinístico)
INSERT INTO gifts (id, wedding_id, name, description, price, image_url, category, status, created_at, updated_at, kind)
SELECT
    'cashttpl-' || w.id,
    w.id,
    'Contribuição em dinheiro',
    '',
    0,
    '',
    'cash',
    'available',
    NOW(),
    NOW(),
    'cash_template'
FROM weddings w
WHERE NOT EXISTS (
    SELECT 1 FROM gifts g WHERE g.wedding_id = w.id AND g.kind = 'cash_template'
);

-- 3. payment_items antes de remover gift_id dos payments (para migração de dados)
CREATE TABLE IF NOT EXISTS payment_items (
    id TEXT PRIMARY KEY,
    payment_id TEXT NOT NULL REFERENCES payments (id) ON DELETE CASCADE,
    gift_id TEXT NOT NULL REFERENCES gifts (id),
    amount DOUBLE PRECISION NOT NULL,
    custom_name TEXT DEFAULT '',
    custom_description TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_items_payment_id ON payment_items (payment_id);
CREATE INDEX IF NOT EXISTS idx_payment_items_gift_id ON payment_items (gift_id);

INSERT INTO payment_items (id, payment_id, gift_id, amount, custom_name, custom_description, created_at)
SELECT
    'pi-' || p.id,
    p.id,
    p.gift_id,
    p.amount,
    '',
    '',
    p.created_at
FROM payments p
WHERE p.gift_id IS NOT NULL;

INSERT INTO payment_items (id, payment_id, gift_id, amount, custom_name, custom_description, created_at)
SELECT
    'pi-' || p.id,
    p.id,
    g.id,
    p.amount,
    '',
    '',
    p.created_at
FROM payments p
JOIN gifts g ON g.wedding_id = p.wedding_id AND g.kind = 'cash_template'
WHERE p.gift_id IS NULL;

ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_gift_id_fkey;
ALTER TABLE payments DROP COLUMN gift_id;
