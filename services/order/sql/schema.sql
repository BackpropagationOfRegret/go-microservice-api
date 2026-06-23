CREATE TABLE IF NOT EXISTS orders (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    restaurant_id TEXT NOT NULL,
    status TEXT NOT NULL,
    total_amount DOUBLE PRECISION NOT NULL,
    items JSONB NOT NULL,
    courier_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
