# Prompt siap paste

Jangan mengetik prompt di panggung — mengetik itu dead air 30 detik yang tidak
menghasilkan apa-apa. Buka file ini, salin, tempel.

Urutannya mengikuti [`DEMO_RUNSHEET.md`](../DEMO_RUNSHEET.md).

---

## 1. `/ce-brainstorm` — time box 12 menit

Seed-nya sengaja rapat: batasan yang **sudah** diputuskan disebutkan di depan, supaya
agent tidak menghabiskan waktu menanyakan hal yang jawabannya sudah ada. Yang disisakan
untuk dia gali hanya bagian yang memang belum diputuskan.

```
/ce-brainstorm Checkout tiket untuk flash sale di repo ini.

Yang sudah ada dan tidak perlu didesain ulang:
- Tabel events dan tickets sudah ada. Satu baris tickets = satu kursi, punya code
  unik per event dan status 'available' | 'sold'.
- Auth JWT sudah jalan, user id tersedia di gin context lewat AuthMiddleware.
- Harga disimpan sebagai price_cents (int64), bukan float.
- Migrasi pakai versioned SQL golang-migrate, bukan AutoMigrate.

Yang diputuskan sekarang, dan HANYA ini:
- Satu endpoint: user membeli satu tiket untuk satu event.
- Kepemilikan tiket setelah terjual: bagaimana dimodelkan.
- Perilaku saat tiket habis.

Di luar ruang lingkup, jangan ditawarkan: refund, waiting list, seat map,
pembayaran, hold/reservasi sementara, pembelian lebih dari satu tiket sekali jalan.

Target: satu story yang bisa diselesaikan dalam satu sesi.
```

**Kalau agent menawarkan scope di luar daftar itu, tolak di tempat.** Refund adalah
story kedua — itu justru bukti compounding-nya nanti.

---

## 2. `/ce-plan` — time box 15 menit

```
/ce-plan Ambil hasil brainstorm barusan dan buat plan implementasi.

Ikuti konvensi repo ini: baca docs/00_AI_CRITICAL_RULES.md dan docs/MODULE_GUIDE.md
dulu. Contoh slice terlengkap ada di `event` (model -> repo -> dto -> service ->
controller -> routes -> mock -> test).

Plan harus menyebut secara eksplisit:
- Migrasi baru yang dibutuhkan, beserta .down.sql-nya.
- Bagaimana kepemilikan tiket dimodelkan.
- Titik mana yang menulis ke database, dan lewat repository yang mana.
- Test apa yang membuktikan fitur ini benar.

Simpan ke docs/plans/checkout.md.
```

---

## 3. `/ce-work` — jendela Q&A besar (22 menit)

```
/ce-work docs/plans/checkout.md
```

`ce-work` punya *shipping tail* — dia bisa commit/push sendiri. Kalau tidak mau itu
terjadi di panggung:

```
/ce-work mode:return-to-caller docs/plans/checkout.md
```

---

## 4. Puncak — load test (15 menit)

Tiga perintah, satu per satu. Jangan pakai `make loadtest` di panggung: exit code-nya
adalah vonis, jadi make menambahkan baris `make: *** Error 1` setelah vonis dan itu
merusak momennya.

```bash
./scripts/loadtest/reset.sh
go run ./scripts/loadtest -n 300 -c 80
```

Lalu, setelah angkanya muncul:

```
/ce-debug Load test barusan menjual 300 tiket dari inventaris 100, dan beberapa kode
tiket terjual ke lebih dari satu pembeli. Cari akar masalahnya, tulis test yang
gagal dulu untuk membuktikannya, baru perbaiki.
```

Setelah fix:

```bash
./scripts/loadtest/reset.sh
go run ./scripts/loadtest -n 300 -c 80
```

---

## 5. `/ce-code-review` — jendela Q&A (10 menit)

```
/ce-code-review
```

Segmen paling lambat: dia menyebar beberapa persona paralel **dan** mengirim kode ke
peer model lain. Ini kandidat pertama yang dipotong kalau waktu mepet.

---

## 6. `/ce-compound` (7 menit)

```
/ce-compound
```

Setelah selesai, tunjukkan buktinya di layar:

```bash
ls docs/solutions/
git log --oneline -3
```

---

## 7. Story kedua — pembuktian compounding (13 menit)

Ini segmen yang menutup argumen. Jangan potong.

```
/ce-work Tambahkan refund satu tiket: tiket yang sudah terjual dikembalikan ke
status available, dan hanya pembelinya yang boleh melakukannya.

Sebelum mulai, baca docs/solutions/.
```

Yang ditunggu: agent memakai pola locking dari story pertama **tanpa disuruh**. Kalau
itu terjadi, tunjukkan barisnya di layar dan bandingkan dengan entri di
`docs/solutions/`. Itu takeaway-nya.

---

## Kalau planning lewat time box

```bash
git checkout demo/plan-ready
```

Bilang apa adanya: *"ini yang tadi mau dia hasilkan, gue potong biar kita sempat lihat
bagian implementasinya."* Audiens tidak akan keberatan. Kehabisan waktu di menit 90
jauh lebih mahal daripada mengakui satu segmen dipotong.
