# Persiapan alat

Diaudit di mesin demo pada 2026-09-11. **Tidak ada yang blocking** — semua yang wajib
sudah terpasang. Yang tersisa hanya keputusan dan pemanasan.

---

## Sudah ada, tidak perlu apa-apa

| Alat | Versi | Dipakai untuk |
|---|---|---|
| Go | 1.25.3 | Semuanya |
| Docker | 27.5.1 (OrbStack) | Postgres, Testcontainers |
| docker compose | v5.1.2 | Stack dev |
| git | 2.49.0 | — |
| gh | 2.74.1 | Kalau mau tunjukkan PR |
| jq | 1.7.1 | Membaca respons JSON di layar |
| Plugin Claude Code | — | `compound-engineering`, `eyay-toolkits` |
| Image `postgres:16-alpine` | — | Compose **dan** Testcontainers |
| Modul `testcontainers-go` | v0.43, v0.44 | Integration test |

Dua yang terakhir penting: keduanya operasi jaringan. Sudah ter-cache, jadi tidak ada
unduhan di panggung.

---

## Satu keputusan yang harus diambil

**`gopls` belum terpasang, dan Serena tidak jalan tanpanya.**

CLAUDE.md global menyebut Serena sebagai alat utama untuk semua pekerjaan kode. Tanpa
`gopls`, setiap kali agent mencoba Serena ia gagal dengan:

```
Found a Go version but gopls is not installed.
```

lalu jatuh ke Read/Edit bawaan. Fungsional, tapi **terlihat sebagai kegagalan di layar**
persis di momen lo ingin menunjukkan alat yang rapi.

Dua pilihan, ambil salah satu sebelum H-1:

**A. Pasang gopls** — Serena hidup, navigasi simbolik jalan.
```bash
go install golang.org/x/tools/gopls@latest
```
Verifikasi: `command -v gopls` mengembalikan path.

**B. Matikan Serena untuk sesi ini** — hilangkan sumber kegagalannya. Lebih aman kalau
lo tidak berencana memamerkan Serena, karena satu MCP yang gagal di awal sesi menurunkan
kepercayaan pada semua yang setelahnya.

Jangan biarkan apa adanya. Gagal-lalu-fallback adalah pilihan terburuk dari ketiganya.

---

## Opsional, tidak dibutuhkan demo

| Alat | Kapan perlu | Pasang |
|---|---|---|
| `migrate` CLI | Hanya untuk `make migrate-down` / `migrate-create`. Migrasi jalan otomatis saat boot, jadi demo tidak menyentuhnya | `brew install golang-migrate` |
| `psql` | Tidak perlu di host — `reset.sh` memakai `docker exec` | `brew install libpq` |
| Image pgadmin | Hanya untuk `make dev` mode penuh | `docker pull dpage/pgadmin4` |

---

## Pemanasan H-1

Jalankan semuanya. Setiap baris di sini pernah gagal di mesin ini sebelum diperbaiki.

```bash
cd oneticket-api

# 1. Dependency terunduh — jangan pernah menunggu ini di depan orang
go mod download

# 2. Postgres. --env-file WAJIB: compose ada di .docker/, tanpa itu
#    ${MASTER_DB_*} kosong dan gagal dengan "invalid proto:"
docker compose --env-file .env -f .docker/docker-compose-dev.yml up -d postgres_db

# 3. Boot sekali: migrasi jalan, event + 100 tiket ter-seed
go run main.go
# tunggu "server listening", lalu Ctrl-C
```

Verifikasi kondisi awal — **keempatnya harus sesuai**:

```bash
# a. Read path hidup, 100 tiket
curl -s localhost:8000/api/v1/events | jq '.data[0] | {name, available_tickets}'
#    -> { "name": "Flash Sale Demo", "available_tickets": 100 }

# b. Checkout BELUM ada — ini kondisi awal yang benar
./scripts/loadtest/reset.sh
go run ./scripts/loadtest -n 300 -c 80
#    -> BELUM ADA YANG TERJUAL (300x 404), exit 3

# c. N+1 MASIH ada — ini bug yang diperbaiki live
go run ./scripts/nplusone
#    -> N+1, 200 event -> 202 query, exit 1

# d. Semua test hijau meski bug (b) dan (c) ada
go test ./tests/... -race
#    -> semua ok, ~15 detik
```

Kalau (c) menjawab `AMAN`, bug-nya sudah keburu diperbaiki — `git log` cari commitnya.

### Cek konfigurasi yang mudah terlewat

```bash
grep -E '^(RATE_LIMIT_RPS|MASTER_DB_HOST|MASTER_DB_PORT|APP_ENV)=' .env
```

Harus:
```
APP_ENV=development
RATE_LIMIT_RPS=1000      <- di 100, rate limiter menolak banjir duluan
MASTER_DB_HOST=localhost <- mode A: API di host
MASTER_DB_PORT=5436
```

`RATE_LIMIT_RPS` adalah yang paling mudah merusak demo tanpa terlihat: di nilai default,
load test kena rate limit sebelum sempat menyentuh logika pembelian.

### Parasut

```bash
git branch -a | grep demo/
#  demo/plan-ready   <- artefak brainstorm + plan, kalau planning lewat time box
#  demo/work-ready   <- checkout terimplementasi, kalau ce-work macet
#  demo/nplusone     <- sudah di-merge ke main
```

Coba sekali: `git checkout demo/work-ready && go test ./tests/... && git checkout main`.
Parasut yang belum pernah dibuka bukan parasut.

---

## Pengaturan layar

- Font terminal besar. Uji dari baris belakang ruangan, bukan dari kursi lo.
- Layar dibagi dua: **agent di kiri, output perintah di kanan**. Penonton harus bisa
  melihat merah-hijau tanpa lo pindah jendela.
- Notifikasi mati. Semua.
- Buka `docs/demo/prompts.md` di jendela terpisah yang bisa lo salin tanpa mengganggu
  dua panel di atas.
- Rekam video cadangan seluruh alur semalam sebelumnya.

## Yang harus dijalankan sekali sebelum tidur

Dry run penuh dengan prompt yang sama persis, catat menit tiap segmen. Agent itu
non-deterministik — yang lo perlukan bukan kepastian, tapi rentang waktunya.
