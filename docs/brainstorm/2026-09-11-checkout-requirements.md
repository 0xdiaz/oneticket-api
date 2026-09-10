---
artifact_contract: ce-product-contract/v1
artifact_readiness: requirements-only
source: ce-brainstorm
date: 2026-09-11
topic: checkout tiket flash sale
plan: docs/plans/2026-09-11-001-feat-checkout-plan.md
---

# Checkout Tiket - Requirements

Dokumen ini menjawab **apa** yang dibangun. **Bagaimana** membangunnya ada di
[`docs/plans/2026-09-11-001-feat-checkout-plan.md`](../plans/2026-09-11-001-feat-checkout-plan.md).

Batas itu disengaja: keputusan produk di sini final dan tidak boleh diubah oleh
planning. Kalau planning menemukan konflik, konfliknya diangkat ke sini, bukan
diselesaikan diam-diam di plan.

---

## Goal Capsule

**Objective.** User yang sudah login bisa membeli satu tiket untuk satu event.
Satu tiket diambil dari stok yang tersedia dan menjadi miliknya, dengan catatan
transaksi yang bisa dilacak.

**Product authority.** Keputusan produk di dokumen ini final. Yang sengaja
dibiarkan terbuka hanya mekanisme teknis pengambilan tiket — lihat R8.

**Open blockers.** Tidak ada.

---

## Masalah

Event dan tiketnya sudah bisa dilihat (`GET /api/v1/events`, `/events/:id`), tapi
tidak ada cara membelinya. Tanpa checkout, inventaris tiket tidak pernah berubah
dan produk belum melakukan apa pun yang bernilai.

Konteks yang membentuk masalah ini: flash sale. Sejumlah kecil tiket, banyak orang
menekan tombol pada detik yang sama.

## Aktor

**A1. Pembeli** — user terdaftar yang sudah login. User id tersedia di gin context
lewat `AuthMiddleware`. Tidak ada aktor lain di scope ini.

## Requirements

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
> mekanisme apa pun untuk memenuhinya — itu keputusan implementasi. Yang mengikat
> adalah testnya, bukan caranya.

## Key Decisions

**KD1 — Kepemilikan dicatat di tabel `orders` terpisah.** Governs R3, R4.
Bukan kolom tambahan di `tickets`. Alasannya: transaksi butuh harga saat beli
(R4), yang tidak punya tempat di baris tiket; dan catatan transaksi yang berdiri
sendiri membuat refund bisa dibangun tanpa membongkar model ini.

**KD2 — Stok habis dijawab `409 Conflict`.** Governs R5.
Konsisten dengan pola error yang sudah ada di repo (`ErrEmailAlreadyExists` juga
409) dan bisa memakai `utils.Conflict` yang sudah tersedia. `410 Gone` lebih tepat
secara semantik tapi menuntut helper baru tanpa manfaat sepadan.

**KD3 — Mekanisme konkurensi tidak ditentukan di sini.** Governs R8.
Menuliskan mekanismenya berarti memutuskan hal teknis di dokumen requirement,
sementara yang benar-benar harus dijamin adalah perilakunya.

## Flow

```
F1. Pembeli (terautentikasi)
  -> POST /api/v1/events/{id}/purchase
      -> event ada?                tidak -> 404
      -> ada tiket available?      tidak -> 409
      -> ambil satu tiket, tandai sold, catat order
  -> 201 Created + detail tiket & order
```

## Acceptance Examples

| # | Kondisi | Hasil |
|---|---|---|
| AE1 | Event punya 100 tiket, pembeli terautentikasi membeli | `201`, tiket jadi `sold`, satu baris `orders` terbentuk |
| AE2 | Semua tiket sudah `sold` | `409`, tidak ada baris `orders` baru |
| AE3 | Tanpa token | `401` |
| AE4 | Event id tidak dikenal | `404` |
| AE5 | Harga event diubah setelah pembelian | Harga di `orders` tidak ikut berubah |
| AE6 | 300 pembelian bersamaan atas 100 tiket | Tepat 100 sukses, 200 ditolak `409`, tidak ada tiket dengan lebih dari satu pemilik |

AE6 adalah acceptance criteria untuk R8, dan harus punya test otomatis yang
membuktikannya — bukan diverifikasi manual.

## Success Criteria

1. Keenam acceptance example punya test yang membuktikannya.
2. `go test ./tests/... -race` hijau, seluruh suite di bawah 3 menit.
3. `api/openapi.yaml` mencakup endpoint baru; spec tetap cocok 100% dengan route
   yang terdaftar.
4. Tidak ada perubahan pada fitur lain.

## Out of Scope

Refund, waiting list, seat map, pembayaran, hold/reservasi sementara, pembelian
lebih dari satu tiket dalam satu permintaan, pemilihan kursi.

Refund akan dibangun sebagai story terpisah setelah ini.

## Constraints

- Ikuti pola slice `event`: model → repository → dto → service → controller →
  routes → mock → test. Rujukan aturan: `docs/00_AI_CRITICAL_RULES.md`,
  `docs/MODULE_GUIDE.md`.
- Perubahan skema hanya lewat versioned SQL golang-migrate, berpasangan
  `.up.sql`/`.down.sql`. Tidak ada AutoMigrate.
- Uang selalu `int64` dalam satuan terkecil. Tidak pernah float di jalur mana pun.
- Integration test memakai harness Testcontainers yang sudah ada di
  `tests/integration/harness`. Tanpa setup manual, tanpa env var.
- Test Driven Development: test ditulis dan dilihat gagal lebih dulu.
