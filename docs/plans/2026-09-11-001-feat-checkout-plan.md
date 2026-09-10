---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
product_contract_source: ce-brainstorm
execution: code
date: 2026-09-11
topic: checkout tiket flash sale
---

# Checkout Tiket - Plan

**Product Contract preservation:** Product Contract tidak berubah. Semua R-ID, KD-ID,
dan AE-ID dipertahankan apa adanya. Enrichment hanya menambah Planning Contract,
Implementation Units, Verification Contract, dan Definition of Done.

## Goal Capsule

**Objective.** User yang sudah login bisa membeli satu tiket untuk satu event.
Satu tiket diambil dari stok yang tersedia dan menjadi miliknya, dengan catatan
transaksi yang bisa dilacak.

**Product authority.** Keputusan produk di dokumen ini sudah final. Yang belum
diputuskan hanya mekanisme teknis pengambilan tiket — itu sengaja diserahkan ke
implementasi, dengan syarat memenuhi acceptance criteria di bawah.

**Open blockers.** Tidak ada.

---

## Product Contract

### Masalah

Event dan tiketnya sudah bisa dilihat (`GET /api/v1/events`, `/events/:id`), tapi
tidak ada cara membelinya. Tanpa checkout, inventaris tiket tidak pernah berubah
dan produk belum melakukan apa pun yang bernilai.

Konteks yang membentuk masalah ini: flash sale. Sejumlah kecil tiket, banyak orang
menekan tombol pada detik yang sama.

### Aktor

**A1. Pembeli** — user terdaftar yang sudah login. User id tersedia di gin context
lewat `AuthMiddleware`. Tidak ada aktor lain di scope ini.

### Requirements

**R1.** Pembeli yang terautentikasi dapat membeli satu tiket untuk satu event.

**R2.** Tiket yang dibeli berpindah dari `available` ke `sold` dan terikat pada
pembelinya.

**R3.** Setiap pembelian menghasilkan catatan transaksi yang menyimpan siapa
membeli tiket mana, untuk event apa, dan berapa harganya saat itu.

**R4.** Harga pada catatan transaksi dibekukan saat pembelian. Perubahan harga
event setelahnya tidak mengubah transaksi yang sudah terjadi.

**R5.** Ketika tidak ada lagi tiket `available`, pembelian ditolak dengan
`409 Conflict`.

**R6.** Permintaan tanpa autentikasi ditolak dengan `401`.

**R7.** Permintaan untuk event yang tidak ada ditolak dengan `404`.

**R8. Tidak boleh terjadi oversell.** Di bawah pembelian yang bersamaan, jumlah
tiket terjual tidak pernah melebihi inventaris, dan satu tiket tidak pernah
menjadi milik lebih dari satu pembeli.

> R8 adalah kriteria hasil, **bukan** resep. Dokumen ini sengaja tidak menyebut
> mekanisme apa pun untuk memenuhinya — itu keputusan implementasi. Yang
> mengikat adalah testnya (U7), bukan caranya.

### Key Decisions

**KD1 — Kepemilikan dicatat di tabel `orders` terpisah.** Governs R3, R4.
Bukan kolom tambahan di `tickets`. Alasannya: transaksi butuh harga saat beli
(R4), yang tidak punya tempat di baris tiket; dan catatan transaksi yang berdiri
sendiri membuat refund bisa dibangun tanpa membongkar model ini.

**KD2 — Stok habis dijawab `409 Conflict`.** Governs R5.
Konsisten dengan pola error yang sudah ada di repo (`ErrEmailAlreadyExists` juga
409) dan bisa memakai `utils.Conflict` yang sudah tersedia.

**KD3 — Mekanisme konkurensi tidak ditentukan di sini.** Governs R8.
Menuliskan mekanismenya berarti memutuskan hal teknis di dokumen requirement,
sementara yang benar-benar harus dijamin adalah perilakunya.

### Flow

```
F1. Pembeli (terautentikasi)
  -> POST /api/v1/events/{id}/purchase
      -> event ada?                tidak -> 404
      -> ada tiket available?      tidak -> 409
      -> ambil satu tiket, tandai sold, catat order
  -> 201 Created + detail tiket & order
```

### Acceptance Examples

| # | Kondisi | Hasil |
|---|---|---|
| AE1 | Event punya 100 tiket, pembeli terautentikasi membeli | `201`, tiket jadi `sold`, satu baris `orders` terbentuk |
| AE2 | Semua tiket sudah `sold` | `409`, tidak ada baris `orders` baru |
| AE3 | Tanpa token | `401` |
| AE4 | Event id tidak dikenal | `404` |
| AE5 | Harga event diubah setelah pembelian | Harga di `orders` tidak ikut berubah |
| AE6 | 300 pembelian bersamaan atas 100 tiket | Tepat 100 sukses, 200 ditolak `409`, tidak ada tiket dengan lebih dari satu pemilik |

### Out of Scope

Refund, waiting list, seat map, pembayaran, hold/reservasi sementara, pembelian
lebih dari satu tiket dalam satu permintaan, pemilihan kursi.

Refund akan dibangun sebagai story terpisah setelah ini.

---

## Planning Contract

### Key Technical Decisions

**KTD1 — Skema `orders`.** Instansiasi KD1.
Kolom: `id`, `user_id`, `event_id`, `ticket_id`, `price_cents` (BIGINT),
`status` (VARCHAR, CHECK `IN ('paid','refunded')`, default `'paid'`),
`created_at`, `updated_at`.
`ticket_id` **UNIQUE** — satu tiket hanya boleh punya satu order. Ini juga
jaring pengaman terakhir untuk R8 di level database.
`status` disertakan sejak awal agar refund tidak perlu membongkar skema.

**KTD2 — `OrderRepository` adalah satu-satunya penulis.** Governs R2, R3.
Pembelian menyentuh dua tabel (`tickets` dan `orders`) dan harus atomik, jadi
seluruh operasi tulis berada di satu method repository, bukan diorkestrasi
service lewat dua panggilan. Service menentukan *kapan*, repository menentukan
*bagaimana*.

**KTD3 — Mekanisme pengambilan tiket ditentukan saat implementasi.** Governs R8.
Yang mengikat: U7 (test konkuren) harus hijau dengan `-race`. Pilihan mekanisme
tidak dibatasi.

**KTD4 — Test konkuren masuk suite default.** 50 goroutine memperebutkan 20
tiket. Cukup untuk memicu race pada implementasi naif, cukup ringan untuk budget
3 menit. Tidak dipisah dengan build tag: regresi race yang hanya ketahuan kalau
seseorang ingat menjalankan suite terpisah bukan jaring pengaman.

### Migrasi

`internal/adapters/database/migrations/sql/000007_create_orders_table.up.sql`
dan pasangan `.down.sql`-nya (`DROP TABLE IF EXISTS orders;`).

Foreign key ke `events(id)` dan `tickets(id)` dengan `ON DELETE RESTRICT` —
order adalah catatan keuangan, tidak boleh hilang diam-diam saat event dihapus.
Index pada `user_id` untuk "order saya".

### Titik tulis ke database

| Operasi | Lewat | Catatan |
|---|---|---|
| Ambil tiket + tandai `sold` + buat order | `OrderRepository.PurchaseTicket` | Satu unit atomik. Satu-satunya penulis pada jalur ini |
| Baca event | `EventRepository.GetByID` | Sudah ada, tidak berubah |
| Hitung ketersediaan | `TicketRepository.CountByStatus` | Sudah ada, tidak berubah |

Tidak ada layer lain yang boleh menyentuh `database.DB` pada jalur ini.

### Perubahan `api/openapi.yaml`

- Path baru `POST /api/v1/events/{id}/purchase`, tag `Events`, `security: bearerAuth`.
- Respons: `201` (`PurchaseResponse`), `401`, `404`, `409`.
- Schema baru `PurchaseResponse`: `order_id`, `ticket_id`, `ticket_code`,
  `event_id`, `price_cents` (integer, int64).
- Spec harus tetap cocok 100% dengan route yang terdaftar.

---

## Implementation Units

Urutan TDD: **U1–U3 menulis test yang gagal lebih dulu.** U4–U6 membuatnya hijau.
U1 dan U2 bisa paralel. U3 menunggu U2 (butuh fake yang sama). U7 dan U8 menunggu
jalur utama hijau.

### U1. Migrasi dan model `Order`

**Goal.** Tabel `orders` ada, dengan constraint yang menegakkan aturannya sendiri.

**Requirements.** R3, R4. Instansiasi KTD1.

**Dependencies.** Tidak ada.

**Files.**
- `internal/adapters/database/migrations/sql/000007_create_orders_table.up.sql`
- `internal/adapters/database/migrations/sql/000007_create_orders_table.down.sql`
- `internal/domain/models/order_model.go`

**Approach.**
1. Tulis pasangan migrasi. Tag struct model harus cocok dengan DDL.
2. `price_cents` BIGINT. Tidak ada tipe pecahan di jalur mana pun.
3. `UNIQUE (ticket_id)` dan `CHECK (status IN ('paid','refunded'))`.
4. Konstanta status di file model, sinkron dengan CHECK — pola sama seperti
   `TicketStatusAvailable` di `ticket_model.go`.

**Patterns to follow.** `internal/domain/models/ticket_model.go` dan migrasi
`000006_create_tickets_table.up.sql`.

**Test scenarios** (`tests/integration/orders/schema_test.go`):
- positive: migrasi naik, tabel `orders` ada dengan seluruh kolom yang diharapkan
- negative: insert dua order untuk `ticket_id` yang sama → ditolak `uq_orders_ticket`
- negative: insert order dengan `status = 'pending'` → ditolak `orders_status_check`
- negative: insert order dengan `ticket_id` yang tidak ada → ditolak foreign key
- edge case: `price_cents = 9007199254740993` bertahan utuh melewati database
  (di atas batas presisi `float64`)

**Acceptance criteria.** `go test ./tests/integration/orders/...` hijau; keempat
constraint terbukti menolak lewat test, bukan lewat pembacaan DDL.

---

### U2. Fake `MockOrderRepository`

**Goal.** Service bisa diuji tanpa database.

**Requirements.** Prasyarat U3.

**Dependencies.** U1 (butuh `models.Order`).

**Files.** `tests/mocks/order_repo_mock.go`

**Approach.**
1. In-memory, `sync.RWMutex`, ikut bentuk `tests/mocks/ticket_repo_mock.go`.
2. Wajib `var _ repositories.OrderRepository = (*MockOrderRepository)(nil)`.
3. Sediakan field error (`PurchaseErr`) agar jalur gagal bisa dipicu tanpa
   menambah fake baru.
4. Sediakan `SoldOut` untuk mensimulasikan stok habis.

**Patterns to follow.** `tests/mocks/ticket_repo_mock.go`.

**Test expectation: none** — fake, bukan unit yang membawa perilaku produk.
Terbukti benar lewat pemakaiannya di U3 dan assertion compile-time.

---

### U3. Test unit `OrderService` (merah dulu)

**Goal.** Perilaku service terkunci sebelum ada implementasinya.

**Requirements.** R1, R5, R6, R7.

**Dependencies.** U2.

**Files.** `tests/unit/services/order_service_test.go`

**Approach.** Black box, `package services_test`, hanya fake dari `tests/mocks/`.
Tidak menyentuh database. Test ini **harus gagal** sebelum U5 ada.

**Test scenarios** — `describe PurchaseTicket`:
- positive: event ada dan ada stok → mengembalikan response berisi
  `ticket_code` dan `price_cents`, memanggil repository tepat sekali
- positive: `price_cents` pada response sama persis dengan harga event saat itu
- negative: event tidak ada → `ErrEventNotFound`
- negative: stok habis → `ErrSoldOut`
- negative: repository gagal → error dibungkus, **bukan** `ErrSoldOut`
  (kegagalan infrastruktur tidak boleh menyamar sebagai stok habis)
- edge case: `price_cents = 9007199254740993` melewati service tanpa berubah

**Acceptance criteria.** Semua skenario di atas ada dan gagal karena
`OrderService` belum ada — bukan karena error kompilasi yang lain.

---

### U4. `OrderRepository`

**Goal.** Satu operasi atomik yang mengambil tiket, menandainya `sold`, dan
mencatat order.

**Requirements.** R2, R3, R8. Instansiasi KTD2.

**Dependencies.** U1.

**Files.** `internal/domain/repositories/order_repo.go`

**Approach.**
1. Interface ter-ekspor + struct unexported + `NewOrderRepository()`, seperti
   `event_repo.go`.
2. Satu method `PurchaseTicket(eventID, userID uint, priceCents int64) (*models.Order, *models.Ticket, error)`.
3. Seluruh operasi tulis dalam satu transaksi. Tidak ada jalur di mana tiket
   berubah `sold` tanpa order tercatat.
4. Stok habis dikembalikan sebagai sentinel error, bukan `gorm.ErrRecordNotFound`.
5. **Mekanisme pengambilan tiket bebas** — lihat KTD3. Yang mengikat U7.

**Patterns to follow.** `internal/domain/repositories/event_repo.go` untuk bentuk;
`refresh_token_repo.go` untuk operasi multi-langkah.

**Test scenarios** (`tests/integration/orders/purchase_repo_test.go`):
- positive: pembelian menandai tepat satu tiket `sold` dan membuat satu order
- positive: order menyimpan `price_cents` yang diberikan, bukan membacanya ulang
- negative: tidak ada tiket `available` → sentinel stok habis, tidak ada order baru
- edge case: tiket terakhir bisa dibeli, tiket berikutnya ditolak
- edge case: kegagalan di tengah tidak meninggalkan tiket `sold` tanpa order

**Acceptance criteria.** Setelah pembelian, `count(tickets WHERE status='sold')`
selalu sama dengan `count(orders)`. Tidak pernah ada selisih.

---

### U5. `OrderService`

**Goal.** Membuat test U3 hijau.

**Requirements.** R1, R5, R7.

**Dependencies.** U3, U4.

**Files.** `internal/app/services/order_service.go`, `internal/app/dto/order_dto.go`

**Approach.**
1. Sentinel error: `ErrSoldOut`. `ErrEventNotFound` dipakai ulang dari
   `event_service.go`, jangan dideklarasikan ulang.
2. Bergantung pada `repositories.EventRepository` dan `repositories.OrderRepository`
   — interface, bukan tipe konkret.
3. Ambil event untuk memvalidasi keberadaannya dan membaca `price_cents`, lalu
   serahkan harga itu ke repository (R4: harga dibekukan di sini).
4. `logger.LogStart`/`LogFinish` dengan span `OrderService.PurchaseTicket`.

**Patterns to follow.** `internal/app/services/event_service.go`.

**Test scenarios.** Sudah ditulis di U3. Unit ini membuatnya hijau tanpa mengubahnya.

**Acceptance criteria.** `go test ./tests/unit/services/...` hijau, dan tidak ada
satu pun skenario U3 yang diubah untuk mencapainya.

---

### U6. Controller, route, dan OpenAPI

**Goal.** Endpoint hidup dan terdokumentasi.

**Requirements.** R1, R5, R6, R7.

**Dependencies.** U5.

**Files.**
- `internal/app/controllers/order_controller.go`
- `internal/app/routers/event_routes.go` (tambah satu route)
- `internal/app/routers/index.go` (wiring)
- `api/openapi.yaml`

**Approach.**
1. `POST /api/v1/events/:id/purchase`, dipasang di grup yang memakai
   `middlewares.AuthMiddleware(authService)` — itu yang memberi R6 secara gratis.
2. `userID := c.GetUint("user_id")`.
3. Pemetaan error: `ErrEventNotFound` → `utils.NotFound`, `ErrSoldOut` →
   `utils.Conflict`, sisanya → `utils.InternalServerError`. Sukses →
   `utils.Created`.
4. Perbarui `api/openapi.yaml` sesuai bagian "Perubahan api/openapi.yaml".

**Patterns to follow.** `internal/app/controllers/event_controller.go`.

**Test scenarios** (`tests/unit/controllers/order_controller_test.go`):
- positive: service sukses → `201`, body memuat `ticket_code`
- negative: tanpa `user_id` di context → `401`
- negative: id bukan angka → `400`
- negative: `ErrEventNotFound` → `404`
- negative: `ErrSoldOut` → `409`
- negative: error tak dikenal → `500`, dan pesan internal tidak bocor ke body

**Acceptance criteria.** Jumlah path di `api/openapi.yaml` sama dengan jumlah
route terdaftar (saat ini 13 → menjadi 14).

---

### U7. Test konkuren — kunci R8

**Goal.** Membuktikan tidak ada oversell, dan menahannya selamanya.

**Requirements.** R8, AE6.

**Dependencies.** U6.

**Files.** `tests/integration/orders/concurrent_purchase_test.go`

**Approach.**
1. Seed satu event dengan **20** tiket.
2. **50** goroutine, semuanya menunggu satu channel gate lalu dilepas bersamaan.
   Race butuh simultanitas, bukan volume — 50 permintaan berurutan tidak akan
   memunculkannya.
3. Kumpulkan hasil, lalu assert.

**Test scenarios** — `describe PurchaseTicket / edge case`:
- Covers AE6: tepat 20 sukses, 30 sisanya gagal dengan sentinel stok habis
- Covers AE6: `count(tickets WHERE status='sold') == 20`
- Covers AE6: jumlah baris `orders` == 20, dan seluruh `ticket_id`-nya unik
- Covers AE6: tidak ada `ticket_id` yang muncul di lebih dari satu order

**Acceptance criteria.** Hijau dengan `-race`. Test ini **harus** merah pada
implementasi read-then-write naif — kalau hijau pada implementasi naif, testnya
yang salah, bukan kodenya.

---

### U8. Test full flow lewat HTTP

**Goal.** Membuktikan seluruh lapisan tersambung, bukan hanya service.

**Requirements.** R1–R7, AE1–AE5.

**Dependencies.** U6.

**Files.** `tests/integration/orders/purchase_flow_test.go`,
`tests/integration/orders/main_test.go`

**Approach.** `main_test.go` berisi `func TestMain(m *testing.M) { os.Exit(harness.RunMain(m)) }`.
Setiap test mulai dengan `harness.Reset(t)`. Jalankan lewat router asli dengan
`httptest`, bukan memanggil service langsung.

**Test scenarios** — `describe POST /api/v1/events/{id}/purchase`:
- positive (Covers AE1): token valid, ada stok → `201`, tiket jadi `sold`,
  satu baris `orders`
- negative (Covers AE3): tanpa header Authorization → `401`
- negative (Covers AE4): event id tidak dikenal → `404`
- negative (Covers AE2): semua tiket `sold` → `409`, jumlah order tidak bertambah
- edge case (Covers AE5): harga event diubah setelah pembelian → `price_cents`
  di order tidak ikut berubah
- edge case: dua pembelian berurutan oleh user yang sama → dua order berbeda
  dengan tiket berbeda

**Acceptance criteria.** Keenam skenario hijau lewat HTTP, tanpa setup manual dan
tanpa env var.

---

## Verification Contract

| Gate | Perintah | Syarat |
|---|---|---|
| Format | `gofmt -l ./internal ./tests` | Kosong |
| Vet | `go vet ./...` | Bersih |
| Build | `go build ./...` | Sukses |
| Unit | `go test ./tests/unit/...` | Hijau |
| Integration | `go test ./tests/integration/...` | Hijau, tanpa setup manual |
| Race | `go test ./tests/... -race` | Hijau |
| Budget | Seluruh suite | Di bawah 3 menit |
| Load test | `./scripts/loadtest/reset.sh && go run ./scripts/loadtest -n 300 -c 80` | `AMAN — terjual 100 dari 100`, exit 0 |

Selama iterasi, jalankan hanya test yang gagal dengan `-run`. Suite penuh cukup
sekali di akhir.

## Definition of Done

1. Kedelapan unit selesai dengan acceptance criteria masing-masing terpenuhi.
2. Seluruh gate Verification Contract hijau.
3. `api/openapi.yaml` cocok 100% dengan route yang terdaftar.
4. Tidak ada perubahan pada fitur lain — `git diff` hanya menyentuh file yang
   terdaftar di unit-unit di atas.
5. Tidak ada tipe pecahan pada jalur uang mana pun.

## Deferred to Follow-Up Work

- Refund (story berikutnya). `orders.status` sudah menyiapkan tempatnya.
- Endpoint "order saya". Index pada `user_id` sudah disiapkan.
- Idempotency key untuk pembelian ganda akibat sinyal buruk.
