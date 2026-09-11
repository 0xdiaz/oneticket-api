# Prompt siap paste, sesi checkout

Jangan mengetik prompt di panggung. Mengetik itu dead air 30 detik yang tidak
menghasilkan apa-apa. Buka file ini, salin, tempel.

Urutan mengikuti [`DEMO_RUNSHEET.md`](../DEMO_RUNSHEET.md).

---

## [1] `/ce-brainstorm`, time box 12 menit

```
/ce-brainstorm Saat ini belum ada fitur checkout: user tidak bisa membeli tiket.
Yang ada baru jalur baca (GET /api/v1/events dan /events/:id).

Flownya: user yang sudah login membeli satu tiket untuk satu event, satu tiket
diambil dari stok yang tersedia dan menjadi miliknya.

Gunakan approach Test Driven Development. Buatkan unit test, integration test,
dan E2E test untuk positive, negative, dan edge case. E2E-nya menumpang pola yang
sudah ada di tests/e2e/.

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
```

**Kalau agent menawarkan scope di luar daftar, tolak di tempat.** Refund itu story
kedua, justru bukti compounding-nya nanti.

---

## [2] `/ce-plan`, time box 15 menit

```
/ce-plan Buat planning dari brainstorm barusan. Simpan ke docs/plans/checkout.md.

Karena ini TDD, plan harus menempatkan task penulisan test SEBELUM task
implementasi, dan menyebut setiap test-nya satu per satu, bukan "tulis test"
sebagai satu task gelondongan.

Atur tasknya agar bisa dikerjakan paralel pakai task Claude Code. Jangan ada code
yang duplicate atau redundant, kalau dua task menyentuh file yang sama, gabungkan
atau beri urutan yang jelas.

Plan harus menyebut eksplisit:
- Acceptance criteria per task, kondisi terukur yang menyatakan task itu selesai.
  Setiap acceptance criteria harus punya test yang membuktikannya.
- Migrasi baru beserta .down.sql-nya.
- Bagaimana kepemilikan tiket dimodelkan.
- Titik mana yang menulis ke database, lewat repository yang mana.
- Perubahan di api/openapi.yaml, termasuk batas nilai dan seluruh status code
  jalur gagal. Spec ini dipakai sebagai alat uji, bukan cuma dokumentasi.
- Daftar test lengkap beserta kategorinya, mengikuti konvensi di bawah.

KONVENSI TEST, plan harus patuh ini:

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
  test. Itu yang bikin lewat budget waktu.
- Harus jalan dengan `go test ./tests/integration/...` tanpa setup manual dan
  tanpa env var. Pattern lama yang membaca TEST_DB_MASTER_DSN lalu t.Skip boleh
  tetap ada untuk test yang sudah ada, tapi jangan dipakai untuk yang baru.

E2E test:
- Lokasi tests/e2e/, ikuti persis pola yang sudah ada di sana. Baca
  tests/e2e/main_test.go, client_test.go, dan journey_test.go dulu.
- Bedanya dengan integration: E2E lewat socket sungguhan, menembus seluruh
  rantai middleware, dan tokennya diterbitkan endpoint login beneran. Tidak
  boleh memanggil service atau repository langsung. Semua assertion lewat
  http.Client.
- Router yang dites wajib routers.SetupRoute(), bukan router yang dirakit
  sendiri di test. Test yang mendaftarkan route-nya sendiri tidak membuktikan
  apa pun soal route yang didaftarkan aplikasi.
- Tambahkan perjalanan checkout ke journey yang sudah ada: daftar, login,
  lihat event, beli tiket, lalu pastikan tiket itu benar jadi miliknya saat
  dibaca ulang lewat HTTP.
- Wajib ada sisi negatifnya: beli tanpa token harus 401, beli event yang tidak
  ada harus 404, beli saat stok habis harus ditolak dengan status yang sudah
  diputuskan di brainstorm.

Smoke test:
- Sudah ada di scripts/smoke/, dijalankan dengan `go run ./scripts/smoke`
  terhadap server yang sudah hidup. Bukan bagian dari `go test`.
- Kalau checkout menambah sesuatu yang wajib hidup setelah deploy, tambahkan
  satu cek ke sana. Ikuti bentuk cek yang sudah ada.
- Cek smoke harus murah dan tidak merusak data. Kalau sebuah cek perlu menulis,
  jangan taruh di smoke.

Negative test harus mencakup semua logic validasi.

Karena ada logic keuangan (price_cents int64), sertakan test precision loss dan
rounding error. Uang tidak boleh pernah jadi float di jalur mana pun.

Wajib ada satu integration test konkuren: N goroutine membeli bersamaan dari
inventaris terbatas, assert tidak ada oversell dan tidak ada tiket yang terjual
dua kali. Dijalankan dengan -race.

Regresi itu aturan, bukan kategori folder. Jangan bikin tests/regression/.
Setiap bug yang diperbaiki wajib berangkat dari satu test yang merah dulu, dan
test itu tetap tinggal di layer asalnya setelah fix. Sebutkan aturan ini di
plan supaya berlaku juga untuk bug yang ketemu saat implementasi.

Kontrak API, Schemathesis:
- api/openapi.yaml itu sumber kebenaran endpoint, dan dipakai sebagai alat uji
  lewat scripts/apitest/run.sh. Spec yang melenceng dari kode langsung terlihat
  di situ, jadi memperbarui spec bukan pekerjaan dokumentasi, itu bagian dari
  test.
- Untuk endpoint checkout yang baru, spec harus menyebut batas nilai yang
  sebenarnya, bukan cuma tipe. Kalau id itu integer positif, tulis minimum-nya.
  Tanpa itu fuzzer akan mengirim angka negatif dan raksasa yang menurut spec sah.
- Semua kemungkinan status code harus terdaftar di spec, termasuk jalur gagal:
  401, 404, dan status saat stok habis.

Keamanan, OWASP ZAP:
- Pesan error yang keluar ke klien tidak boleh membawa detail internal. Tidak
  ada nama driver, nama kolom, tipe SQL, atau teks error dari library. Simpan
  yang detail di log, kirim yang umum ke klien.
- Endpoint checkout wajib menolak sebelum menyentuh database kalau input tidak
  masuk akal, supaya error database tidak pernah jadi jalan keluar.

Seluruh suite tidak boleh lebih dari 3 menit. Kalau perkiraannya lewat, sebutkan
di plan cara menekannya.

Jangan tanya lagi hal yang sudah dijawab di brainstorm.
```

---

## [3] `/ce-work`, jendela Q&A besar, 22 menit

```
/ce-work mode:return-to-caller docs/plans/checkout.md

Kerjakan paralel sesuai plan, dan pertahankan urutan TDD: test ditulis dan
dilihat gagal dulu, baru implementasinya.

Commit kecil-kecil agar mudah direview. Jalankan gofmt, go vet, dan go build
sebelum tiap commit.

Kalau ada test yang gagal, jalankan ulang HANYA test itu (-run) sampai hijau.
Jangan jalankan seluruh suite tiap iterasi, karena itu mahal dan lambat. Suite penuh
dengan -race cukup sekali di akhir untuk memastikan tidak ada regresi.
```

`mode:return-to-caller` menahan *shipping tail*-nya, tanpa itu `ce-work` bisa
commit/push/PR sendiri di tengah demo.

---

## [4] Puncak, tiga ukuran, 15 menit

Tiga alat, tiga pelajaran. Jalankan berurutan. Dua yang pertama masing-masing
lima menit, yang ketiga dua menit, sisanya nafas.

### 4a. Load test, bug yang *dicegah* konteks

```bash
./scripts/loadtest/reset.sh
go run ./scripts/loadtest -n 300 -c 80
```

Hasil yang diharapkan: `AMAN, terjual 100 dari 100 tiket`. Katakan apa adanya:
race condition klasik ini tidak terjadi karena plan menyebut transaksi, unique
constraint, dan test konkuren. Konteks yang menyiapkannya.

### 4b. Query probe, bug yang *lolos semua test*

```bash
go run ./scripts/nplusone
```

Hasil yang diharapkan: `200 event -> 202 query`. Sebelum lanjut, tunjukkan
bahwa seluruh test hijau:

```bash
go test ./tests/... 2>&1 | tail -6
```

Itu poinnya: hasilnya benar, cuma mahal. Tidak ada assertion yang gagal.

Lalu:

```
/ce-debug Satu panggilan GET /api/v1/events menghabiskan 202 query untuk 200
event, satu query tambahan per baris. Seluruh test hijau, jadi ini tidak
tertangkap apa pun.

Ikuti alur ini, jangan langsung lompat ke perbaikan:
1. Reproduce, tulis test otomatis yang gagal dan membuktikan biayanya, bukan
   kebenarannya. Merah dulu.
2. Root analysis, jelaskan penyebabnya, bukan gejalanya.
3. Prioritas, kalau ketemu lebih dari satu masalah, urutkan berdasarkan dampak.
4. Rekomendasi, sebutkan opsi perbaikan beserta trade-off-nya sebelum memilih.
5. Fix sampai hijau.

Test dari langkah 1 harus tetap ada setelah fix sebagai regresi. Selama
iterasi, jalankan hanya test itu (-run), bukan seluruh suite.
```

Setelah fix:

```bash
go run ./scripts/nplusone
```

Target: `AMAN, query tetap 2 meski event naik dari 10 ke 200`.

### 4c. Spec sebagai alat uji, 2 menit

Dua ukuran di atas lo yang menulis alatnya. Yang ini tidak: yang dipakai adalah
`api/openapi.yaml` yang sudah ada di repo sejak awal.

```bash
./scripts/apitest/run.sh
```

Selesai dalam 0,14 detik, dan di kondisi awal repo ia menemukan empat hal di dua
endpoint. Seed-nya dipin di run.sh, jadi hasilnya sama tiap dijalankan. Yang paling
layak ditunjuk:

```
GET /api/v1/events/9223372036854775808  ->  500
    unable to encode ... into binary format for int4 (OID 23)
```

Dua kalimat yang perlu keluar di sini:

1. Spec itu ditulis buat dokumentasi. Diarahkan ke fuzzer, file yang sama jadi
   suite test, dan suite itu tumbuh sendiri tiap spec nambah field.
2. Temuannya bukan soal salah hitung, tapi soal batas yang tidak pernah
   dipikirkan. Tidak ada orang yang akan menulis test `id = 9223372036854775808`.

Tiap temuan datang dengan perintah `curl`-nya, jadi kalau ada yang tidak percaya,
jalankan di layar saat itu juga.

**Jangan diperbaiki live.** Ini bahan untuk segmen berikutnya, bukan story ketiga.

---

## [5] `/ce-code-review` + scan keamanan, jendela Q&A, 10 menit

Segmen ini punya dua hal yang jalan barengan. Mulai yang lambat duluan, di
terminal ketiga, lalu biarkan:

```bash
./scripts/security/run.sh
```

ZAP butuh sekitar satu setengah menit dan membaca `api/openapi.yaml` untuk tahu
endpoint apa saja yang ada, jadi ia tidak menebak dengan cara merayapi. Sementara
ia jalan, mulai review:

```
/ce-code-review Review perubahan di branch ini. Kerjakan paralel.
Outputnya format MD, simpan jadi satu file di docs/reviews/.
```

Sekarang dua hal mengaudit kode yang sama dari dua arah, dan lo punya jendela
Q&A paling lebar di sesi ini. Ini juga segmen paling lambat: persona paralel
**dan** kirim kode ke peer model lain. Kandidat pertama yang dipotong kalau
waktu mepet.

Saat ZAP selesai, yang ditunjuk cukup dua baris:

```
WARN-NEW: A Server Error response code was returned by the server
          /api/v1/events/546058336831160984 (500)
WARN-NEW: X-Content-Type-Options Header Missing  x 6
```

Baris pertama adalah bug yang sama dengan temuan Schemathesis tadi, ditemukan
lewat jalan yang sama sekali berbeda. Itu yang bikin percaya: dua alat yang
tidak saling kenal berhenti di titik yang sama.

Baris kedua tidak akan pernah ketemu oleh test Go mana pun di repo ini, karena
tidak ada handler yang salah. Itu setelan, dan setelan justru yang tidak pernah
dilihat unit test.

Laporan lengkapnya HTML di `scripts/security/reports/`. Buka kalau ada yang
minta, 117 cek lain statusnya PASS dan itu juga informasi.

---

## [6] `/ce-compound`, 7 menit

```
/ce-compound
```

Setelah selesai, tunjukkan buktinya di layar:

```bash
ls docs/solutions/
git log --oneline -5
```

---

## [7] Story kedua, pembuktian compounding, 13 menit

Segmen yang menutup argumen. Jangan potong.

```
/ce-work mode:return-to-caller Tambahkan refund satu tiket: tiket yang sudah
terjual dikembalikan ke status available, dan hanya pembelinya yang boleh
melakukannya.

Sebelum mulai, baca docs/solutions/ dan test yang baru ditulis untuk checkout,
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
