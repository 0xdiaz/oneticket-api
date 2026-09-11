---
artifact_contract: ce-unified-plan/v1
artifact_readiness: requirements-only
product_contract_source: ce-brainstorm
date: 2026-09-11
status: siap untuk ce-plan
---

# Pembelian Tiket (Checkout) - Requirement

## Goal Capsule

**Tujuan.** User yang sudah login bisa membeli **satu** tiket untuk **satu** event lewat satu
request. Tiket diambil dari stok yang tersedia, ditandai terjual, dan kepemilikannya tercatat.

**Otoritas produk.** Dokumen ini. Empat keputusan di bawah sudah ditetapkan dan tidak perlu digali
ulang oleh ce-plan.

**Blocker terbuka.** Satu, lihat [Pertanyaan Terbuka](#pertanyaan-terbuka) — status HTTP untuk
pembelian sebelum masa jual dibuka. Tidak memblokir mayoritas pekerjaan.

**Batas.** Satu story, selesai dalam satu sesi. Test-first.

---

## Masalah

Hanya jalur baca yang ada: `GET /api/v1/events` dan `GET /api/v1/events/{id}`. Katalog bisa
dilihat, tidak bisa dibeli. `tickets` sudah punya satu baris per kursi dengan status
`available | sold`, tapi **tidak punya kolom pemilik** — `ticket_model.go` menyatakan hal itu
sengaja ditunda sampai fitur checkout. `TicketRepository` juga hanya punya jalur baca, dengan
catatan bahwa strategi penguncian sengaja dipilih bersamaan dengan jalur tulis, bukan ditebak
lebih dulu. Story inilah jalur tulis itu.

## Ruang Lingkup

**Termasuk**

- Satu endpoint baru: `POST /api/v1/events/{id}/purchase`, wajib terautentikasi.
- Satu migrasi baru: tabel `purchases`.
- Slice lengkap mengikuti bentuk `event`: model → repository → dto → service → controller →
  routes → mock → test.
- Unit test, integration test (Testcontainers), dan E2E test untuk kasus positif, negatif, dan tepi.
- `api/openapi.yaml` bertambah satu path (13 → 14).

**Tidak termasuk** — jangan ditawarkan, jangan diantisipasi lewat kolom atau abstraksi spekulatif:
refund, pembatalan, waiting list, seat map atau pemilihan kursi, pembayaran, hold/reservasi
sementara, pembelian lebih dari satu tiket dalam satu request, riwayat pembelian
(`GET /purchases`), transfer tiket.

**Tidak boleh tersentuh** — dilindungi kendala demo (`CLAUDE.md`):

- N+1 di `EventService.List` **tetap ada**. Fitur ini tidak mengubah `event_service.go`.
  `go run ./scripts/nplusone` harus tetap melaporkan `200 event -> 202 query`.
- Bug 500 pada `GET /api/v1/events/{id}` untuk id di luar jangkauan `int4` **tetap ada**.
  Endpoint baru menangani id-nya sendiri dengan benar (lihat N7), itu tidak memperbaiki
  yang lama dan tidak boleh dijadikan alasan untuk memperbaikinya.
- Header `X-Content-Type-Options` tetap absen.

---

## Keputusan

### K1 — Kepemilikan lewat tabel `purchases` terpisah

Tabel baru, migrasi berpasangan `000007_create_purchases_table.up.sql` / `.down.sql`.
`tickets` **tidak** mendapat kolom `user_id`. `tickets.status` tetap penanda stok; `purchases`
adalah satu-satunya catatan kepemilikan.

Bentuk yang disepakati:

```sql
CREATE TABLE purchases (
  id           SERIAL PRIMARY KEY,
  ticket_id    INTEGER NOT NULL UNIQUE REFERENCES tickets (id),
  user_id      INTEGER NOT NULL REFERENCES users (id),
  price_cents  BIGINT  NOT NULL CHECK (price_cents >= 0),
  purchased_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

`ticket_id UNIQUE` adalah penjaga terakhir: satu tiket tidak bisa dibeli dua kali, dijaga
database, bukan disiplin kode. Sejalan dengan konvensi repo — constraint hidup di migrasi.

Konsekuensi yang diterima: "siapa pemilik tiket ini" selalu butuh join. Dapat diterima karena
tidak ada jalur baca kepemilikan dalam ruang lingkup.

`price_cents` adalah **snapshot** harga event saat transaksi. Harga event yang berubah kemudian
tidak mengubah baris `purchases` yang sudah ada. `int64` di setiap lapis — DB, model, DTO, JSON.
Tidak ada float di mana pun pada jalur ini.

### K2 — Tidak ada batas pembelian per user

Satu user boleh membeli berulang kali untuk event yang sama lewat request terpisah; setiap
request sukses menghasilkan satu tiket baru. Tidak ada UNIQUE `(event_id, user_id)`.

Alasan yang menentukan: `scripts/loadtest` login sebagai **satu** user lalu menembak N request
paralel. Batas satu-per-user akan membuat probe oversell mustahil gagal, yaitu menghapus
kemampuannya mendeteksi hal yang justru dibuat untuk dideteksi.

### K3 — Stok habis menjawab `409 Conflict`

Service mengembalikan sentinel baru `ErrNoTicketsAvailable`; controller memetakannya lewat
`errors.Is` ke `utils.Conflict`. Mengikuti pola "domain error → HTTP" yang sudah berjalan.
Envelope tetap `{success, message, data, errors}` seperti respons lain.

### K4 — Klaim dengan `FOR UPDATE SKIP LOCKED`

Tiket yang diambil adalah tiket `available` dengan **`id` terkecil**. Klaim memakai
`FOR UPDATE SKIP LOCKED`, sehingga request paralel melewati baris yang sedang dikunci alih-alih
mengantre di belakangnya.

**Konsekuensi dari K1 yang harus dipegang ce-plan:** karena kepemilikan tinggal di tabel
terpisah, klaim bukan lagi satu statement tunggal. Ia dua statement —
`UPDATE tickets ... RETURNING` lalu `INSERT INTO purchases` — dan **keduanya wajib berada dalam
satu transaksi**. Tiket yang berubah menjadi `sold` tanpa baris `purchases` pasangannya adalah
stok yang hilang selamanya dan tidak terdeteksi oleh test mana pun yang hanya melihat status.

Nol baris terpengaruh pada `UPDATE` berarti stok habis → K3.

Pemilihan lapisan mana yang memegang transaksi diserahkan ke ce-plan, dengan satu kendala dari
konvensi repo: hanya lapisan repository yang boleh menyentuh `database.DB`, dan tidak ada
`*gorm.DB` yang di-inject.

---

## Kontrak API

```
POST /api/v1/events/{id}/purchase
Authorization: Bearer <access_token>       (wajib)
Request body: tidak ada
```

Tidak adanya body bukan kelalaian: itu yang membuat "beli lebih dari satu tiket" mustahil
diminta secara struktural, bukan sekadar ditolak validasi.

**201 Created**, `data`:

| Field | Tipe | Catatan |
|---|---|---|
| `id` | `uint` | id baris `purchases` |
| `event_id` | `uint` | |
| `ticket_id` | `uint` | |
| `ticket_code` | `string` | kode kursi, unik dalam satu event |
| `price_cents` | `int64` | snapshot harga, integer sampai ke klien |
| `purchased_at` | `time.Time` | |

`user_id` **tidak** dikembalikan. Pemiliknya adalah pemanggil; mengembalikannya tidak menambah
informasi apa pun bagi klien.

**Identitas selalu dari token.** `user_id` diambil dari `gin.Context` key `"user_id"` yang
dipasang `AuthMiddleware`, tidak pernah dari body, query, atau header lain. Ini yang menutup
IDOR pada endpoint ini — tidak ada jalan bagi pemanggil untuk membeli atas nama orang lain.

**Rate limit.** Grup `/api/v1` sudah memakai `RateLimitMiddleware`. Endpoint ini mewarisinya.
Tidak ada kontrol tambahan.

**openapi.yaml.** Satu path baru, 13 → 14. Spec ditulis tangan dan saat ini cocok 100% dengan
route terdaftar; angka itu harus tetap cocok setelah story ini.

---

## Aturan Perilaku

Ini yang harus merah lebih dulu.

### Positif

| # | Perilaku |
|---|---|
| P1 | User terautentikasi membeli pada event yang punya stok → `201`, satu baris `purchases` baru, satu tiket berubah `available` → `sold`. |
| P2 | `available_tickets` pada `GET /api/v1/events/{id}` turun tepat 1 setelah satu pembelian sukses. |
| P3 | Tiket yang dipilih adalah `available` dengan `id` terkecil saat tidak ada kontensi. |
| P4 | User yang sama membeli dua kali berturut-turut → dua tiket **berbeda**, dua baris `purchases`. |
| P5 | `purchases.price_cents` sama dengan `events.price_cents` pada saat pembelian. |

### Negatif

| # | Perilaku |
|---|---|
| N1 | Tanpa header `Authorization` → `401`, `success=false`. |
| N2 | Token sampah / terpotong / kedaluwarsa → `401`. |
| N3 | Event tidak ada → `404` (`ErrEventNotFound`, sentinel yang sudah ada). |
| N4 | Semua tiket event sudah `sold` → `409` (`ErrNoTicketsAvailable`). |
| N5 | Masa jual belum dibuka (`sale_starts_at` di masa depan) → ditolak. Status: lihat Pertanyaan Terbuka. |
| N6 | `{id}` bukan angka (`/events/abc/purchase`) → `400`. |
| N7 | `{id}` di luar jangkauan `int4` (mis. `99999999999999`) → `400`, **bukan** `500`, dan tanpa membocorkan pesan driver. Endpoint baru menangani ini sendiri; `GET /events/{id}` yang lama tetap dibiarkan rusak. |

Tidak satu pun respons gagal boleh membocorkan pesan database, SQL, atau nama kolom.

### Tepi

| # | Perilaku |
|---|---|
| E1 | Event dengan nol tiket `available` sejak awal → `409` pada request pertama. |
| E2 | 200 request paralel terhadap event berisi 100 tiket → tepat 100 × `201` dan 100 × `409`. Nol oversell. |
| E3 | Dua request paralel dari user yang sama → dua tiket berbeda, bukan satu tiket yang dihitung dua kali. |
| E4 | `price_cents = 9007199254740993` bertahan utuh melewati DB, service, DTO, dan JSON. Di atas batas integer eksak `float64`, jadi konversi float di titik mana pun membuatnya merah. |
| E5 | Harga event diubah setelah pembelian → baris `purchases` lama tidak ikut berubah. |
| E6 | Tidak ada baris `tickets` berstatus `sold` tanpa baris `purchases` pasangannya, termasuk setelah beban konkuren. |

---

## Pembagian Lapisan Test

Pembagian ini mengikuti `CLAUDE.md`; ce-plan menentukan nama file, bukan lapisannya.

- **Unit** (`tests/unit/services`, `package services_test`) — service dengan repository palsu.
  Memetakan P1, P5, N3, N4, E4. Fake baru wajib membawa penegasan waktu-kompilasi
  `var _ repositories.PurchaseRepository = (*MockPurchaseRepository)(nil)`.
- **Integration** (`tests/integration/...`, harness Testcontainers) — Postgres asli. Memegang
  semua yang hanya bisa dibuktikan database: P3, E2, E3, E5, E6, dan constraint
  `ticket_id UNIQUE`. `SKIP LOCKED` dan atomisitas transaksi tidak punya arti di depan mock.
- **E2E** (`tests/e2e/`, menumpang pola yang ada) — satu perjalanan lewat socket nyata dan
  seluruh rantai middleware: register → login → `GET /events` → `POST /purchase` →
  `GET /events/{id}` dan lihat stok berkurang. Plus setengah negatifnya: `POST /purchase`
  tanpa token menjawab `401`. Pakai `routers.SetupRoute()`, bukan router yang dirakit di dalam
  test — lapisan inilah satu-satunya yang menangkap route benar yang terpasang di grup salah.

Probe `scripts/loadtest` bukan test dan tidak menggantikan E2; ia verifikasi terpisah terhadap
server hidup. Defaultnya sudah menunjuk `POST /api/v1/events/{id}/purchase`, jadi kontrak di atas
membuatnya langsung jalan tanpa flag.

---

## Asumsi

Tercatat karena tidak dikonfirmasi eksplisit; koreksi di sini lebih murah daripada di kode.

1. `purchased_at` memakai `DEFAULT NOW()` di database, bukan waktu yang dikirim aplikasi.
2. Migrasi `down` cukup `DROP TABLE purchases`. Tidak ada kolom yang ditambahkan ke `tickets`,
   jadi tidak ada yang perlu dibatalkan di sana.
3. Seeder dev tidak diubah. "Flash Sale Demo" dengan 100 tiket sudah cukup untuk E2 dan loadtest.
4. Tidak ada endpoint baca kepemilikan, sehingga tidak ada indeks pada `purchases.user_id`
   dalam story ini. Menambahkannya sekarang adalah kontrol spekulatif.
5. Event yang sudah lewat (`starts_at` di masa lalu) tetap boleh dibeli. Tidak ada aturan yang
   melarangnya hari ini dan tidak ada permintaan untuk menambahkannya.

## Pertanyaan Terbuka

1. **Status HTTP untuk N5** (beli sebelum `sale_starts_at`). `events.sale_starts_at` dan
   `EventResponse.SaleOpen` sudah ada, jadi kasusnya nyata dan harus ditolak — yang belum pasti
   hanya kodenya. Kandidat: `409 Conflict` dengan pesan berbeda dari stok habis (konsisten dengan
   K3, satu bentuk penolakan untuk satu jenis masalah), atau `403 Forbidden`. Rekomendasi: `409`.
   Putuskan saat ce-plan; tidak memblokir pekerjaan lain.

## Definisi Selesai

- Setiap perilaku di tabel P/N/E punya test yang merah sebelum implementasi dan hijau sesudahnya.
- `gofmt -w . && go vet ./... && go build ./... && go test ./tests/... -race` bersih.
- `go run ./scripts/nplusone` tetap melaporkan `200 event -> 202 query`.
- `make loadtest` melaporkan nol oversell.
- `api/openapi.yaml` berisi 14 path dan cocok dengan route yang terdaftar.
- Tidak ada file melebihi 300 baris, tidak ada fungsi melebihi 100 baris.
- Commit dipecah di sambungan alami: migrasi + model, repository + fake, test, service,
  lalu routes + controller + openapi.
