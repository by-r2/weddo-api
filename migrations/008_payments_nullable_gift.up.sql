-- Contribuições em dinheiro não estão ligadas a uma linha em `gifts`.
ALTER TABLE payments ALTER COLUMN gift_id DROP NOT NULL;
