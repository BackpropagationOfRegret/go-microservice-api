CREATE TABLE IF NOT EXISTS restaurants (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    is_open BOOLEAN NOT NULL DEFAULT TRUE,
    open_from TEXT NOT NULL DEFAULT '09:00',
    open_to TEXT NOT NULL DEFAULT '22:00'
);

CREATE TABLE IF NOT EXISTS menu_items (
    id TEXT PRIMARY KEY,
    restaurant_id TEXT NOT NULL REFERENCES restaurants (id),
    name TEXT NOT NULL,
    description TEXT,
    price DOUBLE PRECISION NOT NULL,
    available BOOLEAN NOT NULL DEFAULT TRUE
);
