-- Create events table
-- Migration: 000005_create_events_table
-- Created to match internal/domain/models/event_model.go

CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    venue VARCHAR(200) NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    sale_starts_at TIMESTAMPTZ NOT NULL,
    -- Price is stored in the smallest currency unit as an integer.
    -- Never use FLOAT/DOUBLE for money.
    price_cents BIGINT NOT NULL CHECK (price_cents >= 0),
    -- Planned inventory: how many ticket rows were generated for this event.
    -- Live availability is derived by counting tickets, not by decrementing this.
    total_tickets INTEGER NOT NULL CHECK (total_tickets > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_sale_starts_at ON events (sale_starts_at);

COMMENT ON TABLE events IS 'Events whose tickets are sold';
COMMENT ON COLUMN events.price_cents IS 'Ticket price in the smallest currency unit (integer, never float)';
COMMENT ON COLUMN events.total_tickets IS 'Planned inventory; live availability is counted from the tickets table';
