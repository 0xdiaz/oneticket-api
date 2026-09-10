-- Create tickets table
-- Migration: 000006_create_tickets_table
-- Created to match internal/domain/models/ticket_model.go

CREATE TABLE IF NOT EXISTS tickets (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    code VARCHAR(32) NOT NULL,
    status VARCHAR(16) NOT NULL CHECK (status IN ('available', 'sold')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- A ticket code identifies a seat within one event, not globally.
    CONSTRAINT uq_tickets_event_code UNIQUE (event_id, code)
);

-- Supports "find the available tickets for this event", the hot path of a sale.
CREATE INDEX IF NOT EXISTS idx_tickets_event_status ON tickets (event_id, status);

COMMENT ON TABLE tickets IS 'Individual sellable tickets, one row per seat';
COMMENT ON COLUMN tickets.status IS 'available | sold';
