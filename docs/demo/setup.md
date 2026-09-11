# Persiapan alat

Diaudit di mesin demo pada 2026-09-11. **Tidak ada yang blocking.** Semua yang wajib
sudah terpasang. Yang tersisa hanya keputusan dan pemanasan.

---

## Sudah ada, tidak perlu apa-apa

| Alat | Versi | Dipakai untuk |
|---|---|---|
| Go | 1.25.3 | Semuanya |
| Docker | 27.5.1 (OrbStack) | Postgres, Testcontainers |
| docker compose | v5.1.2 | Stack dev |
| git | 2.49.0 | - |
| gh | 2.74.1 | Kalau mau tunjukkan PR |
| jq | 1.7.1 | Membaca respons JSON di layar |
| Plugin Claude Code | - | `compound-engineering`, `eyay-toolkits` |
| Image `postgres:16-alpine` | - | Compose **dan** Testcontainers |
| Modul `testcontainers-go` | v0.43, v0.44 | Integration test, E2E test |
| `uv` / `uvx` | 0.11.32 | Menjalankan Schemathesis tanpa install permanen |
| Schemathesis | 4.26.1 | Fuzzing kontrak API, sudah ter-cache di `~/.cache/uv` |
| Image `ghcr.io/zaproxy/zaproxy:stable` | 1.08 GB | Scan keamanan |

Lima yang terakhir penting: semuanya operasi jaringan. Sudah ter-cache, jadi tidak ada
unduhan di panggung.

Schemathesis sengaja tidak dipasang permanen. `scripts/apitest/run.sh` memanggilnya
lewat `uvx`, yang menariknya ke environment sekali pakai. Repo ini tetap tidak punya
dependency Python, dan mesin lo tidak ketambahan apa-apa.

---

## MCP: matikan semuanya untuk sesi demo

Sudah diputuskan: **Serena dimatikan.** Tanpa `gopls` ia gagal di setiap pemanggilan
lalu jatuh ke Read/Edit. Terlihat sebagai kegagalan di layar persis saat lo ingin
menunjukkan alat yang rapi.

Menghapusnya lewat `claude mcp remove` bukan jawabannya: Serena terdaftar di **user
scope** (`~/.claude.json`), jadi menghapusnya kena semua repo lo.

Yang dipakai: `--strict-mcp-config` dengan config kosong. Non-destruktif, setup global
lo tidak disentuh, dan cukup tutup terminal untuk kembali normal.

```bash
claude --strict-mcp-config --mcp-config .claude/demo-mcp.json
```

`.claude/demo-mcp.json` sudah ada di repo, isinya `{"mcpServers": {}}`.

**Diverifikasi empiris**, bukan diasumsikan:

```
tanpa flag  : "apakah kamu punya mcp__serena__find_symbol?" -> YA
dengan flag : "apakah kamu punya mcp__serena__find_symbol?" -> TIDAK
```

Catatan: `claude mcp list` **tidak** menghormati flag ini, ia tetap menampilkan daftar
yang terkonfigurasi. Jangan pakai itu untuk memverifikasi; pakai pertanyaan di atas.

### Efek samping yang justru diinginkan

Tanpa flag, sesi lo memuat delapan MCP server dan **tiga di antaranya bermasalah**:

| Server | Status |
|---|---|
| `plugin:github:github` | ✘ Gagal, `Authorization header is badly formatted` |
| `claude.ai Tavily` | ! Butuh autentikasi |
| `claude.ai Xero` | ! Butuh autentikasi |

Ketiganya memunculkan peringatan di awal sesi. Tiga baris error sebelum lo mengetik apa
pun adalah pembukaan yang buruk untuk sesi yang tesisnya soal perkakas yang rapi. Flag
ini menghilangkan semuanya sekaligus.

Yang ikut hilang dan memang tidak dibutuhkan demo: Figma, Google Drive, Playwright
(tidak ada frontend), dan context7. Kalau lo ingin context7 tetap hidup, misalnya
berjaga kalau agent perlu mencari dokumentasi GORM, isi `.claude/demo-mcp.json` dengan
entri context7 saja alih-alih objek kosong.

### Yang TIDAK ikut mati

**Review lintas model lewat Codex tetap jalan.** `ce-code-review` memanggil peer lewat
CLI (`codex-cli 0.145.0`), bukan lewat MCP, jadi flag ini tidak menyentuhnya. Ini
penting karena review lintas model adalah salah satu poin sesi lo.

Skill dan plugin juga tidak terpengaruh, flag ini hanya soal MCP.

---

## Plugin dan skill yang terpasang

| Marketplace | Skill | Dipakai demo |
|---|---|---|
| `compound-engineering-plugin` | 37 | **Ya**, enam skill di bawah |
| `eyay-toolkits` | 38 | Tidak (BMad, blockchain, design-thinking) |
| `claude-plugins-official` | 31 | Tidak |

Enam yang dipakai, semuanya sudah diverifikasi ada:

```
/ce-brainstorm   /ce-plan   /ce-work
/ce-debug        /ce-code-review   /ce-compound
```

106 skill terpasang dan lo cuma memakai enam. Itu wajar dan tidak perlu dibereskan, karena
skill tidak dimuat sampai dipanggil. Yang perlu diingat cuma satu: **jangan mengetik
`/bmad-sprint-run`** (dari `eyay-toolkits`). Definisinya menjalankan semua story di
semua epic tanpa jeda sampai selesai atau blocked, sekali diketik, kendali atas jam
dinding hilang.

---

## Opsional, tidak dibutuhkan demo

| Alat | Kapan perlu | Pasang |
|---|---|---|
| `migrate` CLI | Hanya untuk `make migrate-down` / `migrate-create`. Migrasi jalan otomatis saat boot, jadi demo tidak menyentuhnya | `brew install golang-migrate` |
| `psql` | Tidak perlu di host, `reset.sh` memakai `docker exec` | `brew install libpq` |
| Image pgadmin | Hanya untuk `make dev` mode penuh | `docker pull dpage/pgadmin4` |

---

## Pemanasan H-1

Jalankan semuanya. Setiap baris di sini pernah gagal di mesin ini sebelum diperbaiki.

```bash
cd oneticket-api

# 0. Sesi demo dijalankan dengan MCP dimatikan, lihat bagian MCP di atas
#    claude --strict-mcp-config --mcp-config .claude/demo-mcp.json

# 1. Dependency terunduh, jangan pernah menunggu ini di depan orang
go mod download

# 2. Postgres. --env-file WAJIB: compose ada di .docker/, tanpa itu
#    ${MASTER_DB_*} kosong dan gagal dengan "invalid proto:"
docker compose --env-file .env -f .docker/docker-compose-dev.yml up -d postgres_db

# 3. Boot sekali: migrasi jalan, event + 100 tiket ter-seed
go run main.go
# tunggu "server listening", lalu Ctrl-C
```

Verifikasi kondisi awal, **keempatnya harus sesuai**:

```bash
# a. Read path hidup, 100 tiket
curl -s localhost:8000/api/v1/events | jq '.data[0] | {name, available_tickets}'
#    -> { "name": "Flash Sale Demo", "available_tickets": 100 }

# b. Checkout BELUM ada, ini kondisi awal yang benar
./scripts/loadtest/reset.sh
go run ./scripts/loadtest -n 300 -c 80
#    -> BELUM ADA YANG TERJUAL (300x 404), exit 3

# c. N+1 MASIH ada, ini bug yang diperbaiki live
go run ./scripts/nplusone
#    -> N+1, 200 event -> 202 query, exit 1

# d. Semua test hijau meski bug (b) dan (c) ada
go test ./tests/... -race
#    -> semua ok, ~20 detik, tests/e2e ikut di dalamnya

# e. Smoke test terhadap server yang hidup
go run ./scripts/smoke
#    -> SEHAT, 9 dari 9 cek lolos

# f. Kontrak API. Di kondisi awal ini WAJIB menemukan sesuatu
./scripts/apitest/run.sh
#    -> 4 failures, salah satunya 500 di /events/{id}, exit bukan 0

# g. Scan keamanan, paling lambat, sekitar 1,5 menit
./scripts/security/run.sh
#    -> WARN-NEW: 2, FAIL-NEW: 0, PASS: 117
```

Kalau (c) menjawab `AMAN`, bug-nya sudah keburu diperbaiki, `git log` cari commitnya.

Kalau (f) menjawab `AMAN`, seseorang sudah memperbaiki bug 500-nya. Itu justru
merusak segmen 4c, karena bahannya hilang. Cek `git log -- internal/app/controllers`.

### Jebakan versi migrasi, ini pernah bikin `main` tidak bisa boot

Dry run `demo/work-ready` menjalankan migrasi ke-7 di database dev. `main` cuma
punya enam. Setelah balik ke `main`, boot gagal dengan `no migration found for
version 7`, dan gejalanya tidak menyebut branch sama sekali.

Periksa sebelum tidur:

```bash
docker exec oneticket_pg_db psql -U oneticket -d oneticket -c \
  'SELECT version, dirty FROM schema_migrations;'
```

Di `main` harus `6` dan `dirty = f`. Kalau `7`, kembalikan:

```bash
docker exec oneticket_pg_db psql -U oneticket -d oneticket -c \
  'DROP TABLE IF EXISTS orders CASCADE; UPDATE schema_migrations SET version = 6;'
```

Lakukan ini tiap kali selesai mencoba parasut `demo/work-ready`, termasuk saat
latihan. Ini kegagalan yang paling mahal di daftar ini, karena terjadi di menit
nol dan penyebabnya tidak kelihatan dari pesan errornya.

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
non-deterministik. Yang lo perlukan rentang waktunya, bukan kepastian.
