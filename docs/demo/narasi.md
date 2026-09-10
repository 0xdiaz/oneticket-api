# Narasi demo

Bukan naskah kata per kata — itu terdengar dibacakan. Ini **beat**: apa yang harus
mendarat di tiap segmen, kalimat yang menahannya, dan apa yang dikatakan sambil agent
bekerja.

Urutan mengikuti [`DEMO_RUNSHEET.md`](../DEMO_RUNSHEET.md). Prompt siap salin ada di
[`prompts.md`](prompts.md). Jawaban untuk pertanyaan agent ada di
[`answers.md`](answers.md).

---

## Kalimat yang harus dibawa pulang

Kalau penonton cuma ingat satu hal:

> **Agentic coding gagal bukan karena modelnya bodoh, tapi karena konteksnya kosong.**

Setiap segmen harus menopang itu. Kalau ada bagian yang tidak, potong.

Dan satu kalimat lagi yang muncul di akhir, setelah dibuktikan:

> **Yang di-compound bukan kodenya. Pelajarannya.**

---

## 00:00 — Pembukaan (8 menit)

**Tujuan.** Pasang ekspektasi, dan cegah dua salah paham sebelum sempat terbentuk.

Salah paham pertama: orang mengira mereka akan melihat swarm agent otonom. Pakai empat
tingkat itu untuk menempatkan sesi ini:

```
Claude Code single agent       -> mengendalikan 1 agent
Background agent               -> agent otomatis, bisa mereview agent
Human orchestrator multi-agent -> manusia mengendalikan banyak agent
AI orchestrator multi-agent    -> agent mengendalikan agent
```

> "Hari ini kita di tingkat satu, sedikit menyentuh dua. Bukan karena tingkat empat
> mustahil, tapi karena yang menentukan hasil bukan jumlah agentnya."

Salah paham kedua: orang mengira demo AI selalu diatur. Dahului tuduhannya:

> "Nanti ada satu bug yang muncul di depan kalian. Itu bukan bug yang saya tanam —
> implementasi yang wajar memang menghasilkan bug itu, dan saya akan tunjukkan kenapa
> test-nya semua hijau saat bug itu ada."

Lalu peta pipeline, dan tunjuk bahwa **semua kotaknya dijalankan hari ini**:

```
Brainstorm -> Planning -> Work + Verifikasi -> Review -> Compound
```

> "Bukan potongan. Satu putaran penuh, dari belum ada fitur sampai pelajarannya
> tercatat."

---

## 00:08 — Kondisi awal repo (8 menit)

**Tujuan.** Ini segmen yang paling sering dilewatkan, dan yang paling menentukan apakah
tesis lo dipercaya. Kalau penonton tidak melihat konteksnya dulu, mereka akan mengira
agentnya sakti.

Tunjukkan, jangan ceritakan:

```bash
wc -l CLAUDE.md docs/00_AI_CRITICAL_RULES.md docs/MODULE_GUIDE.md
ls docs/
```

> "Ini yang dibaca agent sebelum menulis baris pertama. Bukan prompt saya yang bagus —
> reponya yang sudah menjelaskan dirinya sendiri."

Buka `docs/00_AI_CRITICAL_RULES.md`, scroll bagian *Quick Decision Tree*. Lalu buka satu
slice nyata — `internal/app/services/event_service.go` — dan tunjukkan komentar ini:

```go
// Availability is counted per event, so this issues one query per event on
// top of the list query. Fine at demo scale, and deliberately left that way:
// it is the honest starting point for a later performance pass.
```

> "Perhatikan baris ini. Saya sengaja tinggalkan, dan sengaja dikomentari. Ingat ini,
> nanti kita balik ke sini."

**Itu menanam puncaknya.** Penonton akan mengenali momennya sendiri nanti — jauh lebih
kuat daripada lo yang mengumumkannya.

Tutup dengan kejujuran yang membangun kepercayaan:

> "Repo ini tidak selalu begini. Waktu saya mulai, dokumentasinya menggambarkan
> arsitektur yang tidak ada di kodenya — 15 file, sepuluh ribu baris, semuanya
> menyuruh agent membangun di direktori yang tidak pernah dibuat. Yang kalian lihat
> ini hasil membereskan itu dulu."

---

## 00:16 — `/ce-brainstorm` (12 menit) · time box keras

**Tujuan.** Menunjukkan bahwa agent yang bagus **bertanya**, dan pertanyaannya bagus.

Sebelum paste, katakan apa yang akan terjadi:

> "Saya sudah tahu apa yang mau dibangun. Yang saya belum putuskan cuma tiga hal. Lihat
> apakah agent ini menanyakan tiga itu, atau malah menawarkan fitur yang tidak saya
> minta."

Paste prompt [1]. Sambil dia berpikir:

> "Prompt ini panjang bukan karena saya suka mengetik. Semua yang **sudah** diputuskan
> saya sebut di depan, supaya dia tidak menghabiskan giliran menanyakan hal yang
> jawabannya sudah ada. Yang saya sisakan cuma tiga."

**Ketika pertanyaan muncul, baca jawabannya dari `answers.md`.** Jangan berpikir di
panggung — itu lambat dan terlihat ragu.

Pertanyaan ketiga adalah yang penting. Jawabannya:

> "Itu keputusan implementasi. Yang saya kunci hasilnya — tidak boleh oversell — plus
> test yang membuktikannya."

Lalu jelaskan kenapa, karena ini pelajaran tersendiri:

> "Kalau saya jawab 'pakai row lock', saya baru saja memutuskan hal teknis di dokumen
> requirement. Requirement itu soal apa yang harus benar, bukan bagaimana caranya."

**Kalau lewat 12 menit:** hentikan, `git checkout demo/plan-ready`, bilang apa adanya.

---

## 00:28 — `/ce-plan` (15 menit) · time box keras

**Tujuan.** Menunjukkan bahwa "konteks" itu artefak konkret, bukan kata-kata motivasi.

Paste prompt [2]. Sambil menunggu, ini slot terbaik untuk poin TDD:

> "Perhatikan yang saya minta: task menulis test harus **sebelum** task implementasi,
> dan tiap test disebut satu per satu. Kalau tidak diminta begitu, yang keluar biasanya
> satu task bernama 'write tests' di paling bawah — dan itu bukan TDD, itu test yang
> ditempel belakangan."

Setelah plan jadi, buka dan tunjuk tabel Requirements Trace:

> "Tiap requirement punya unit yang memenuhinya dan test yang membuktikannya. Baris tanpa
> unit berarti plan-nya belum selesai. Jadi tabel ini bukan dokumentasi — itu alat cek."

---

## 00:43 — `/ce-work` (22 menit) · jendela Q&A terbesar

**Tujuan.** Ini bagian yang paling sedikit butuh narasi dan paling banyak butuh Q&A.

Paste prompt [3]. Katakan sekali di depan:

> "Sekarang dia kerja. Ini akan makan dua puluh menitan, jadi mari kita pakai waktunya —
> tanya apa saja."

**Kalau tidak ada yang bertanya**, pancing dengan salah satu dari empat ini. Yang
terakhir paling ampuh:

1. "Ini bedanya apa sama autocomplete di IDE?"
2. "Kalau agentnya salah dan nggak ketahuan gimana?"
3. "Ini habis berapa token sekali jalan?"
4. **"Tim saya nggak punya dokumentasi sebagus ini. Mulai dari mana?"**

Jawaban untuk yang keempat, karena ini pertanyaan yang paling sering ditanya dan paling
berguna dijawab dengan baik:

> "Jangan mulai dari menulis dokumentasi. Mulai dari satu slice yang benar — satu fitur,
> ditulis serapi mungkin, dengan test yang jujur. Lalu tunjuk itu di prompt: 'ikuti pola
> di sini'. Satu contoh nyata lebih kuat daripada lima ribu baris standar. Dokumentasi
> menyusul dari situ, bukan mendahuluinya."

**Minta satu orang memberi kode isyarat waktu agent berhenti.** Lo tidak akan sadar
sendiri karena sedang bicara.

---

## 01:05 — Puncak (15 menit)

Dua alat, dua pelajaran. Ini bagian yang menentukan apakah sesi lo diingat.

### 4a — Bug yang *dicegah*

```bash
./scripts/loadtest/reset.sh
go run ./scripts/loadtest -n 300 -c 80
```

Layar: `AMAN — terjual 100 dari 100 tiket`.

Di sini ada godaan untuk buru-buru lewat. Jangan. **Ini poin pertama lo, dan tidak
terlihat seperti poin kalau tidak dijelaskan:**

> "Ini race condition klasik. Seratus tiket, tiga ratus orang menekan tombol bersamaan.
> Implementasi naif akan menjual tiga ratus. Kalian baru saja melihatnya tidak terjadi.
>
> Bukan karena modelnya pintar. Karena plan-nya menyebut transaksi, menyebut unique
> constraint, dan yang paling menentukan — test konkuren itu ada di plan, terlihat oleh
> agent sebelum dia menulis implementasinya. Dia tahu akan diuji seperti apa."

Tunjukkan barisnya:

```bash
grep -A3 'SKIP LOCKED' internal/domain/repositories/order_repo.go
```

> "Konteks yang bagus tidak membuat agent lebih pintar. Konteks yang bagus membuat
> jawaban yang benar jadi jawaban yang paling jelas."

### 4b — Bug yang *lolos semua test*

Sekarang balik ke komentar yang lo tanam di menit 8.

```bash
go run ./scripts/nplusone
```

Layar: `200 event -> 202 query`.

Berhenti sebentar. Lalu:

```bash
go test ./tests/... 2>&1 | tail -6
```

Semua hijau.

> "Dua ratus dua query untuk satu request. Dan seluruh test hijau.
>
> Karena hasilnya **benar**. Cuma mahal. Tidak ada assertion yang gagal, tidak ada error
> di log, tidak ada yang merah. Test menguji kebenaran, bukan biaya.
>
> Ini kelas bug yang berbeda dari yang tadi. Yang tadi dicegah oleh konteks yang bagus.
> Yang ini tidak dicegah apa pun — dan tidak akan ketahuan sampai ada yang mengukur."

Paste prompt `/ce-debug`. Sambil dia bekerja:

> "Perhatikan saya tidak menyuruh dia memperbaiki. Saya menyuruh dia mereproduksi dulu —
> test yang gagal karena biayanya, bukan karena hasilnya — lalu menyebutkan opsi
> perbaikan beserta trade-off-nya sebelum memilih. Bagian menimbang itu yang biasanya
> tidak kelihatan."

Setelah fix:

```bash
go run ./scripts/nplusone
```

`AMAN — query tetap 2 meski event naik dari 10 ke 200`.

> "Dua ratus dua jadi dua."

Biarkan angka itu menggantung sebentar. Jangan langsung lanjut.

---

## 01:20 — `/ce-code-review` (10 menit) · jendela Q&A

**Tujuan.** Menunjukkan review yang berjalan paralel dan lintas model.

> "Yang jalan sekarang beberapa persona sekaligus, plus satu peer di model lain.
> Alasannya sederhana: model yang menulis kode adalah pembaca terburuk untuk kode itu.
> Dia sudah yakin kodenya benar — dia baru saja meyakinkan dirinya sendiri."

Kandidat pertama yang dipotong kalau waktu mepet. Kalau dipotong, cukup katakan apa yang
biasanya dia temukan.

---

## 01:30 — `/ce-compound` (7 menit)

**Tujuan.** Menunjukkan mekanisme yang jadi nama seluruh pendekatan ini.

Setelah selesai:

```bash
ls docs/solutions/
git log --oneline -5
```

> "Pelajarannya tadi sekarang ada di repo, dan sudah ter-commit. Bukan di kepala saya,
> bukan di catatan pribadi, bukan di Slack yang hilang dalam dua minggu."

> "Ini bagian yang biasanya dilewat, dan ini justru satu-satunya bagian yang membuat
> putaran berikutnya lebih murah dari putaran ini."

---

## 01:37 — Story kedua (13 menit) · jendela Q&A

**Tujuan.** Membuktikan klaimnya. Jangan potong segmen ini.

Paste prompt [7]. Sebelum agent mulai, buat prediksi yang bisa jatuh:

> "Saya tidak akan menyebut locking sama sekali di prompt ini. Kalau nanti dia memakai
> pola yang sama dengan story pertama, itu karena dia membaca `docs/solutions/` — bukan
> karena saya menyuruhnya."

Membuat prediksi di depan itu berisiko, dan justru itu yang membuatnya bernilai. Kalau
meleset, katakan meleset — penonton akan lebih percaya semua yang lain.

Kalau tepat, tunjukkan berdampingan: entri `docs/solutions/` di kiri, kode baru di kanan.

> "Ini yang dimaksud compound. Bukan kodenya yang menumpuk. Pelajarannya."

---

## 01:50 — Penutup (8 menit)

Empat hal, jangan lebih:

1. **Konteks > prompt.** Yang membuat outputnya konsisten adalah repo yang menjelaskan
   dirinya, bukan prompt yang panjang.
2. **Agent harus bisa mengukur.** Test membuat dia mandiri. Tapi test menguji kebenaran —
   untuk biaya dan konkurensi, butuh alat ukur sendiri.
3. **Learning di-commit.** Kesalahan yang tidak dicatat akan terjadi lagi di story
   berikutnya.
4. **Review tetap manusia.** Yang berubah levelnya: dari mengecek sintaks jadi mengecek
   keputusan.

Lalu tutup dengan yang paling jujur, karena ini yang membedakan sesi lo dari konten
"AI bikin saya 10x lebih cepat":

> "Yang kalian lihat hari ini bukan agentnya yang hebat. Sebagian besar kerja saya minggu
> ini bukan menulis prompt — tapi membereskan dokumentasi supaya cocok dengan kodenya,
> membetulkan tooling yang rusak, dan membuat dua alat ukur. Setelah itu, agentnya jadi
> berguna.
>
> Urutannya penting. Kalau repo kalian belum menjelaskan dirinya sendiri, menambah agent
> cuma mempercepat pembuatan kekacauan."

---

## Kalau ada yang tidak berjalan

| Situasi | Katakan |
|---|---|
| Agent nyasar | "Nah, ini yang tadi saya bilang. Kita lihat dia keluar dari sini." Maksimal 3 attempt |
| Sudah 3 attempt | `git checkout demo/final`. "Saya potong. Yang tadi salah karena X." |
| Probe tidak menunjukkan N+1 | Naikkan `-sizes 10,500`. Kemiringannya yang penting |
| Load test justru oversell | **Lebih bagus.** Pakai itu sebagai puncak, N+1 jadi babak kedua |
| Q&A habis, agent masih jalan | Bahas kenapa doc harus cocok dengan kode. Cerita 15 file yang salah |
| Kehabisan waktu | Potong `ce-code-review` dulu. Jangan potong story kedua |

**Prinsip yang berlaku di semua baris di atas:** kesalahan yang diakui menaikkan
kepercayaan, kesalahan yang ditutupi menurunkannya. Penonton yang melihat lo memotong
satu segmen dengan jujur akan lebih percaya pada segmen yang berhasil.
