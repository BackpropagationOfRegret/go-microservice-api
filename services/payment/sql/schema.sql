CREATE TABLE IF NOT EXISTS payments (
    id TEXT PRIMARY KEY,
    order_id TEXT UNIQUE NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    method TEXT NOT NULL DEFAULT 'card',
    status TEXT NOT NULL,
    transaction_id TEXT,
    refund_transaction_id TEXT,
    refund_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE payments ADD COLUMN IF NOT EXISTS refund_transaction_id TEXT;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS refund_reason TEXT;
