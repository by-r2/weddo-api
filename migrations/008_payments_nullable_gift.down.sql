-- Remove contribuições em dinheiro antes de tornar gift_id obrigatório de novo.
DELETE FROM payments WHERE gift_id IS NULL;
ALTER TABLE payments ALTER COLUMN gift_id SET NOT NULL;
