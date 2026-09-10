# Prompt siap paste — sesi checkout

Jangan mengetik prompt di panggung. Mengetik itu dead air 30 detik yang tidak
menghasilkan apa-apa. Buka file ini, salin, tempel.

Urutan mengikuti [`DEMO_RUNSHEET.md`](../DEMO_RUNSHEET.md).

---

## [1] `/ce-brainstorm` — time box 12 menit

```
/ce-brainstorm Saat ini belum ada fitur checkout: user tidak bisa membeli tiket.
Yang ada baru jalur baca (GET /api/v1/events dan /events/:id).

Flownya: user yang sudah login membeli satu tiket untuk satu event, satu tiket
diambil dari stok yang tersedia dan menjadi miliknya.

Gunakan approach Test Driven Development. Buatkan unit test dan integration test
untuk positive, negative, dan edge case.

Constraint:
- Pelajari pattern source code yang ada dulu, buat konsisten. Slice terlengkap ada
  di `event`: model -> repository -> dto -> service -> controller -> routes -> mock
  -> test. Baca juga docs/00_AI_CRITICAL_RULES.md dan docs/MODULE_GUIDE.md.
- Jangan ubah fitur lain. Menambah kolom atau tabel baru lewat migrasi baru boleh.
- Migrasi wajib versioned SQL golang-migrate berpasangan .up.sql/.down.sql.
  Tidak ada AutoMigrate.
- Uang selalu integer (price_cents, int64). Tidak boleh float.
- Update api/openapi.yaml untuk endpoint baru. Spec saat ini cocok 100% dengan
  route yang terdaftar, jangan sampai melenceng.
- Integration test pakai Testcontainers (Postgres asli di container, bukan mock,
  bukan env var). Test harus jalan tanpa setup manual.

Sudah diputuskan, tidak perlu digali lagi:
- Tabel events dan tickets sudah ada. Satu baris tickets = satu kursi, punya code
  unik per event dan status 'available' | 'sold'.
- Auth JWT sudah jalan, user id ada di gin context lewat AuthMiddleware.

Yang perlu diputuskan, dan HANYA ini:
1. Kepemilikan tiket setelah terjual dimodelkan bagaimana.
2. Perilaku saat tiket habis.
3. Cara mengambil tiket yang tersedia.

Di luar ruang lingkup, jangan ditawarkan: refund, waiting list, seat map,
pembayaran, hold sementara, beli lebih dari satu tiket sekali jalan.

Batasi ke maksimal 5 pertanyaan, tanyakan sekaligus dalam satu giliran, dan hanya
untuk tiga poin di atas. Target: satu story yang selesai dalam satu sesi.

Simpan hasilnya ke docs/brainstorm/ sebagai dokumen requirement — apa yang
dibangun saja, tanpa rencana implementasi. Rencana implementasi urusan ce-plan.
```

**Kalau agent menawarkan scope di luar daftar, tolak di tempat.** Refund itu story
kedua — justru bukti compounding-nya nanti.

---

## [2] `/ce-plan` — time box 15 menit

```
/ce-plan Buat planning dari dokumen requirement di docs/brainstorm/ barusan.
Simpan ke docs/plans/.

Jangan salin ulang requirement, aktor, flow, atau acceptance example ke dalam
plan — rujuk lewat ID dan sertakan tabel Requirements Trace yang memetakan tiap
R-ID ke unit yang memenuhinya dan test yang membuktikannya. Dua salinan
requirement akan saling menyimpang.

Karena ini TDD, plan harus menempatkan task penulisan test SEBELUM task
implementasi, dan menyebut setiap test-nya satu per satu — bukan "tulis test"
sebagai satu task gelondongan.

Atur tasknya agar bisa dikerjakan paralel pakai task Claude Code. Jangan ada code
yang duplicate atau redundant — kalau dua task menyentuh file yang sama, gabungkan
atau beri urutan yang jelas.

Plan harus menyebut eksplisit:
- Acceptance criteria per task — kondisi terukur yang menyatakan task itu selesai.
  Setiap acceptance criteria harus punya test yang membuktikannya.
- Migrasi baru beserta .down.sql-nya.
- Bagaimana kepemilikan tiket dimodelkan.
- Titik mana yang menulis ke database, lewat repository yang mana.
- Perubahan di api/openapi.yaml.
- Daftar test lengkap beserta kategorinya, mengikuti konvensi di bawah.

KONVENSI TEST — plan harus patuh ini:

Penamaan: describe method -> describe positive/negative/edge case -> test.

Unit test:
- Lokasi tests/unit/<layer>/, package <layer>_test (black box).
- Isolated ala London school: pakai fake dari tests/mocks/, tidak menyentuh DB.
- Fake wajib punya assertion var _ repositories.XRepository = (*MockX)(nil).

Integration test:
- Lokasi tests/integration/, harus ada full flow.
- Pakai Testcontainers: Postgres asli di container, migrasi dijalankan ke
  container itu, lalu test jalan di atasnya.
- DB tidak boleh di-mock. Pihak ketiga (kalau nanti ada) semua di-mock.
- Container dibagi satu per package lewat TestMain, jangan satu container per
  test — itu yang bikin lewat budget waktu.
- Harus jalan dengan `go test ./tests/integration/...` tanpa setup manual dan
  tanpa env var. Pattern lama yang membaca TEST_DB_MASTER_DSN lalu t.Skip boleh
  tetap ada untuk test yang sudah ada, tapi jangan dipakai untuk yang baru.

Negative test harus mencakup semua logic validasi.

Karena ada logic keuangan (price_cents int64), sertakan test precision loss dan
rounding error. Uang tidak boleh pernah jadi float di jalur mana pun.

Wajib ada satu integration test konkuren: N goroutine membeli bersamaan dari
inventaris terbatas, assert tidak ada oversell dan tidak ada tiket yang terjual
dua kali. Dijalankan dengan -race.

Seluruh suite tidak boleh lebih dari 3 menit. Kalau perkiraannya lewat, sebutkan
di plan cara menekannya.

Jangan tanya lagi hal yang sudah dijawab di brainstorm.
```

---

## [3] `/ce-work` — jendela Q&A besar, 22 menit

```
/ce-work mode:return-to-caller docs/plans/checkout.md

Kerjakan paralel sesuai plan, dan pertahankan urutan TDD: test ditulis dan
dilihat gagal dulu, baru implementasinya.

Commit kecil-kecil agar mudah direview. Jalankan gofmt, go vet, dan go build
sebelum tiap commit.

Kalau ada test yang gagal, jalankan ulang HANYA test itu (-run) sampai hijau.
Jangan jalankan seluruh suite tiap iterasi — itu mahal dan lambat. Suite penuh
dengan -race cukup sekali di akhir untuk memastikan tidak ada regresi.
```

`mode:return-to-caller` menahan *shipping tail*-nya — tanpa itu `ce-work` bisa
commit/push/PR sendiri di tengah demo.

---

## [4] Puncak — load test, 15 menit

Tiga perintah, satu per satu. **Jangan** pakai `make loadtest`: exit code-nya adalah
vonis, jadi make menambahkan baris `make: *** Error 1` tepat setelah vonis muncul dan
itu merusak momennya.

```bash
./scripts/loadtest/reset.sh
go run ./scripts/loadtest -n 300 -c 80
```

Setelah angkanya muncul:

```
/ce-debug Load test barusan menjual 300 tiket dari inventaris 100, dan beberapa
kode tiket terjual ke lebih dari satu pembeli.

Ikuti alur ini, jangan langsung lompat ke perbaikan:
1. Reproduce — tulis test otomatis yang gagal dan membuktikan bugnya. Merah dulu.
2. Root analysis — jelaskan penyebabnya, bukan gejalanya.
3. Prioritas — kalau ketemu lebih dari satu masalah, urutkan berdasarkan dampak.
4. Rekomendasi — sebutkan opsi perbaikan beserta trade-off-nya sebelum memilih.
5. Fix sampai hijau.

Test dari langkah 1 harus tetap ada setelah fix sebagai regresi. Selama
iterasi, jalankan hanya test itu (-run), bukan seluruh suite.
```

Setelah fix:

```bash
./scripts/loadtest/reset.sh
go run ./scripts/loadtest -n 300 -c 80
```

---

## [5] `/ce-code-review` — jendela Q&A, 10 menit

```
/ce-code-review Review perubahan di branch ini. Kerjakan paralel.
Outputnya format MD, simpan jadi satu file di docs/reviews/.
```

Segmen paling lambat: menyebar beberapa persona paralel **dan** mengirim kode ke peer
model lain. Kandidat pertama yang dipotong kalau waktu mepet.

---

## [6] `/ce-compound` — 7 menit

```
/ce-compound
```

Setelah selesai, tunjukkan buktinya di layar:

```bash
ls docs/solutions/
git log --oneline -5
```

---

## [7] Story kedua — pembuktian compounding, 13 menit

Segmen yang menutup argumen. Jangan potong.

```
/ce-work mode:return-to-caller Tambahkan refund satu tiket: tiket yang sudah
terjual dikembalikan ke status available, dan hanya pembelinya yang boleh
melakukannya.

Sebelum mulai, baca docs/solutions/ dan test yang baru ditulis untuk checkout —
ikuti konvensi yang sama persis dari situ.

TDD: test dulu sampai gagal, baru implementasi. Wajib ada positive, negative
untuk seluruh logic validasi, edge case, dan satu integration test konkuren
dengan -race.

Commit kecil-kecil.
```

Yang ditunggu: agent memakai pola locking dari story pertama **tanpa disuruh**. Kalau
itu terjadi, tunjukkan barisnya dan bandingkan dengan entri di `docs/solutions/`. Itu
takeaway-nya.

---

## Kalau planning lewat time box

```bash
git checkout demo/plan-ready
```

Bilang apa adanya: *"ini yang tadi mau dia hasilkan, gue potong biar kita sempat lihat
bagian implementasinya."* Audiens tidak akan keberatan. Kehabisan waktu di menit 90
jauh lebih mahal daripada mengakui satu segmen dipotong.
