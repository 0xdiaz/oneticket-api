# Flash sale load test

Fires concurrent purchase requests at a running server and says, in one line,
whether the event oversold. Exits non-zero when it did, so the result is a
pass/fail rather than a wall of numbers.

## Why it exists

The checkout endpoint is **not implemented** in this repository. It is built
live during the sharing session (see [`docs/DEMO_RUNSHEET.md`](../../docs/DEMO_RUNSHEET.md)).
This tool is written to survive that: it works before checkout exists, against
a naive first implementation, and against the fixed one.

| State | Output | Exit |
|---|---|---|
| No checkout endpoint yet | `BELUM ADA YANG TERJUAL` (404s) | 3 |
| Naive read-then-write | `OVERSELL, terjual 300 dari 100 tiket` | 1 |
| Correct locking | `AMAN, terjual 100 dari 100 tiket` | 0 |

All three were verified before this was committed.

## Usage

```bash
# 1. Postgres up, server running (see the repo README for the two run modes)
docker compose --env-file .env -f .docker/docker-compose-dev.yml up -d postgres_db
go run main.go

# 2. Reset the event to a full house
./scripts/loadtest/reset.sh

# 3. Attack
go run ./scripts/loadtest -n 300 -c 80
```

Reset between every attempt. The tool reports availability before and after,
so a half-sold event makes the numbers meaningless.

## Flags

| Flag | Default | Notes |
|---|---|---|
| `-n` | 500 | total purchase requests |
| `-c` | 100 | how many in flight at once |
| `-path` | `/api/v1/events/{id}/purchase` | change if the live implementation picks another route |
| `-method` | `POST` | |
| `-body` | `{}` | request body |
| `-event` | 1 | event id |
| `-token` | *(empty)* | Bearer token; when empty it logs in as `-email` |
| `-email` | `admin@example.local` | seeded demo user |
| `-url` | `http://localhost:8000` | |

## How it decides

It is deliberately tolerant of a contract that has not been designed yet:

- **Any 2xx counts as a sold ticket.** No assumption about the response shape.
- **A ticket code is extracted when present**, trying `ticket_code`, `code`,
  `ticket`, `seat`, `seat_code` at any nesting depth. That is what catches the
  same seat going to two buyers. If the response carries no code, duplicate
  detection quietly switches off and the other checks still work.
- **Availability is read from `GET /api/v1/events/{id}` before and after**,
  which works no matter how checkout is implemented.

A run is an oversell when any of these hold: more 2xx than tickets existed,
more tickets consumed than existed, availability went negative, or one code
went to more than one buyer.

## Notes for the demo

- Every worker blocks on one gate and is released together. A race needs
  simultaneity, not just volume: trickling 300 requests in sequence will not
  reproduce it.
- `RATE_LIMIT_RPS` in `.env` is set high on purpose. At the default 100 the
  rate limiter rejects the flood first and the oversell never appears, which
  demos the wrong thing.
- If the live implementation lands on a different route, pass `-path`. Nothing
  else needs changing.
- `make loadtest` does reset + attack in one step, but because the exit code is
  the verdict, make appends its own `make: *** [loadtest] Error 1` line after a
  failing run. On stage, prefer running the two commands separately so the
  verdict is the last thing on screen.
