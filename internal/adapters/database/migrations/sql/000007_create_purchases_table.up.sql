-- Create purchases table
-- Migration: 000007_create_purchases_table
-- Created to match internal/domain/models/purchase_model.go

CREATE TABLE IF NOT EXISTS purchases (
    id SERIAL PRIMARY KEY,

    -- One purchase per ticket, enforced by the database rather than by
    -- application discipline. This is the last guard against selling the same
    -- seat twice: even a claim path with a race writes at most one row here.
    ticket_id INTEGER NOT NULL UNIQUE REFERENCES tickets (id),

    user_id INTEGER NOT NULL REFERENCES users (id),

    -- Snapshot of the event price at the moment of sale. Repricing the event
    -- later must not rewrite what someone paid, so this is a copy rather than
    -- a lookup. Smallest currency unit as an integer; never FLOAT/DOUBLE.
    price_cents BIGINT NOT NULL CHECK (price_cents >= 0),

    purchased_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE purchases IS 'Ownership record for a sold ticket; tickets.status tracks stock, this tracks who holds the seat';
COMMENT ON COLUMN purchases.ticket_id IS 'The seat bought; UNIQUE so one ticket cannot be sold twice';
COMMENT ON COLUMN purchases.price_cents IS 'Price paid in the smallest currency unit (integer, never float), snapshotted at sale time';
