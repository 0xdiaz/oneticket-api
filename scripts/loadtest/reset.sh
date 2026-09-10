#!/usr/bin/env bash
# Reset the demo event back to a full house so the load test starts from a
# known state. Run this between attempts.
#
#   ./scripts/loadtest/reset.sh
#
# Talks to the dev Postgres container directly, so it works even when the
# checkout path has left rows in a broken state.
set -euo pipefail

CONTAINER="${PG_CONTAINER:-oneticket_pg_db}"
DB_USER="${MASTER_DB_USER:-oneticket}"
DB_NAME="${MASTER_DB_NAME:-oneticket}"

if ! docker ps --format '{{.Names}}' | grep -qx "$CONTAINER"; then
  echo "Container '$CONTAINER' tidak jalan." >&2
  echo "Jalankan dulu: docker compose --env-file .env -f .docker/docker-compose-dev.yml up -d postgres_db" >&2
  exit 1
fi

# Drop anything the checkout flow created, if those tables exist yet, then put
# every ticket back to available. The DO block keeps this working before the
# orders table has been designed.
docker exec -i "$CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -q <<'SQL'
DO $$
BEGIN
  IF to_regclass('public.orders') IS NOT NULL THEN
    EXECUTE 'TRUNCATE TABLE orders CASCADE';
  END IF;
END
$$;

UPDATE tickets SET status = 'available' WHERE status <> 'available';
SQL

docker exec -i "$CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -t -c \
  "SELECT '  tiket tersedia: ' || count(*) FROM tickets WHERE status = 'available';"
