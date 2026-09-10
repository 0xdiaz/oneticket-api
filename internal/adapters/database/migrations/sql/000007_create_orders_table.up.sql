-- Create orders table
-- Migration: 000007_create_orders_table
-- Created to match internal/domain/models/order_model.go

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    event_id INTEGER NOT NULL REFERENCES events (id) ON DELETE RESTRICT,
    ticket_id INTEGER NOT NULL REFERENCES tickets (id) ON DELETE RESTRICT,

    -- Price is frozen at purchase time, in the smallest currency unit.
    -- Never FLOAT/DOUBLE: an order is a financial record.
    price_cents BIGINT NOT NULL CHECK (price_cents >= 0),

    status VARCHAR(16) NOT NULL DEFAULT 'paid' CHECK (status IN ('paid', 'refunded')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One ticket can be sold exactly once. This is the last line of defence
    -- against overselling: even if application code races, the database
    -- refuses the second order for the same seat.
    CONSTRAINT uq_orders_ticket UNIQUE (ticket_id)
);

-- Supports "my orders".
CREATE INDEX IF NOT EXISTS idx_orders_user ON orders (user_id);

COMMENT ON TABLE orders IS 'Record of a completed ticket purchase';
COMMENT ON COLUMN orders.price_cents IS 'Price at purchase time in the smallest currency unit (integer, never float)';
COMMENT ON COLUMN orders.status IS 'paid | refunded';
