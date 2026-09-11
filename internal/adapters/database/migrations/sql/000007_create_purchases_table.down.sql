-- Rollback purchases table creation
-- Migration: 000007_create_purchases_table
--
-- No columns were added to tickets, so there is nothing to undo there.

DROP TABLE IF EXISTS purchases;
