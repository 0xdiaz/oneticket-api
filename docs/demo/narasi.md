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

## Kenapa bukan multi-agent

Pertanyaan ini hampir pasti datang, dan jawabannya memperkuat tesis lo — jadi jangan
defensif. Sebut sendiri di pembukaan, satu kalimat setelah empat tingkat agent:

> "Tingkat tiga dan empat itu ada, dan di Claude Code namanya agent teams. Hari ini
> sengaja tidak dipakai. Nanti saya jelaskan kenapa, dan alasannya bukan karena belum
> sempat."

Lalu saat ditanya, tiga alasan. Yang ketiga yang paling penting.

**Satu — bentuk pekerjaannya tidak cocok.** Dokumentasinya sendiri menulis: *"For
sequential tasks, same-file edits, or work with many dependencies, a single session or
subagents are more effective."* Implementasi di demo ini persis itu: unit yang saling
bergantung, file yang sama. Teams unggul untuk riset paralel, review, dan debat
hipotesis — bukan implementasi berurutan.

**Dua — menyalakannya mengubah hal lain.** *"A subagent that Claude names launches as a
teammate, so teams can form even when you didn't ask for one."* `ce-code-review` memang
menyebar subagent bernama; begitu teams menyala, roster itu berubah jadi teammate. Fitur
eksperimental yang diam-diam mengubah bagian pipeline yang sudah diuji bukan hal yang
dinyalakan seminggu sebelum tampil.

**Tiga — lima Claude punya blind spot yang sama.** Ini yang benar-benar penting:

> "Teammate itu instance Claude Code. Model yang sama, cara salah yang sama, cuma lebih
> banyak. Lima agent yang saling berdebat tetap lima agent dengan asumsi yang mirip.
>
> Yang saya butuhkan bukan lebih banyak agent — tapi satu yang **beda cara salahnya**.
> Makanya review di sini lewat model lain, Codex, bukan lewat lebih banyak Claude."

Paralelisme menambah kecepatan. Model yang berbeda menambah **sudut pandang**. Untuk
review, yang kedua jauh lebih berharga.

---

## Bank jawaban Q&A

Jawaban yang sudah disiapkan untuk pertanyaan yang kemungkinan besar datang. Ambil
intinya, jangan hafalkan kalimatnya.

### "Bedanya apa sama Copilot atau autocomplete di IDE?"

> "Autocomplete melanjutkan kalimat yang sedang saya tulis. Yang tadi kalian lihat itu
> membaca plan, menulis migrasi, menulis test, melihatnya gagal, lalu menulis
> implementasinya — dan mengukur hasilnya sendiri.
>
> Perbedaan yang sebenarnya bukan di modelnya. Autocomplete tidak punya konsep
> 'selesai'. Agent punya: ada definition of done, ada gate verifikasi, dan dia tahu dia
> belum selesai sampai gate itu hijau."

### "Berapa biayanya sekali jalan?"

Jangan mengarang angka. Yang jujur:

> "Saya tidak punya angka pasti untuk sesi ini karena belum saya ukur per-run. Yang bisa
> saya katakan soal bentuk biayanya: review paling mahal — dia menyebar beberapa persona
> plus satu model lain. Implementasi jauh lebih murah dari yang orang kira, karena
> sebagian besar tokennya dipakai membaca, bukan menulis.
>
> Dan biaya yang lebih penting bukan token. Kerja minggu ini yang paling banyak makan
> waktu bukan menjalankan agent — tapi membereskan repo sampai agentnya berguna."

### "Kalau agentnya salah dan nggak ketahuan gimana?"

Ini pertanyaan terbaik yang bisa datang. Jawab dengan contoh nyata dari repo ini:

> "Itu terjadi di repo ini, dan bukan kasus kecil.
>
> Endpoint forgot-password mengembalikan token reset password di body respons — siapa
> pun yang tahu email korban bisa ambil akunnya dalam dua request. Test suite-nya
> **hijau**. Bukan cuma hijau: ada test yang secara eksplisit memastikan token itu
> dikembalikan. Jadi siapa pun yang memperbaikinya akan melihat test merah dan mengira
> dirinya yang salah.
>
> Jadi jawabannya: test hijau itu bukti yang lebih lemah dari yang kita kira. Makanya
> di sesi ini saya tidak cuma menjalankan test — saya mengukur. Load test untuk
> konkurensi, query probe untuk biaya. Dua-duanya menangkap hal yang test tidak
> tangkap."

### "Ini bisa dipakai di codebase legacy yang berantakan?"

> "Bisa, tapi urutannya kebalik dari yang orang kira.
>
> Repo ini waktu saya mulai punya lima belas dokumen yang menggambarkan arsitektur yang
> tidak ada di kodenya. Dokumen yang paling awal dibaca agent justru menyuruh membangun
> di direktori yang tidak pernah dibuat. Agent yang nurut ke dokumentasi itu akan
> menghasilkan kode yang yatim piatu di tengah codebase.
>
> Jadi untuk legacy: jangan mulai dengan menyuruh agent mengerjakan fitur. Mulai dengan
> menyuruhnya **membaca satu area dan melaporkan apa yang sebenarnya ada di sana**, lalu
> perbaiki dokumennya. Itu pekerjaan yang justru cocok untuk agent, dan hasilnya modal
> untuk semua pekerjaan setelahnya."

### "Apa yang agent tetap tidak bisa?"

> "Dia tidak bisa tahu apa yang seharusnya benar kalau tidak ada yang memberitahunya.
>
> Contoh dari hari ini: ada test yang saya minta dia tulis untuk event dengan nol tiket.
> Database menolaknya, karena ada CHECK constraint yang melarang event tanpa inventaris.
> Testnya yang salah, bukan kodenya. Yang mengoreksi bukan agent, bukan saya — tapi
> database yang punya aturannya.
>
> Itu pola umumnya. Agent sangat baik mengikuti aturan yang tertulis, dan buta terhadap
> aturan yang cuma ada di kepala seseorang. Makanya aturan yang penting harus ditaruh di
> tempat yang gagal dengan berisik: constraint database, compile error, test — bukan
> kalimat di dokumen."

### "Kode kita dikirim ke mana? Aman?"

Jawab lurus, jangan berkelit:

> "Untuk review, iya — ada satu pass yang dikirim ke model lain, dan skill-nya memang
> mengumumkan itu sebelum jalan. Repo yang saya pakai hari ini publik, jadi tidak ada
> masalah.
>
> Untuk repo kerjaan, itu keputusan kebijakan, bukan keputusan teknis. Passnya bisa
> dimatikan, dan ada fallback reviewer lokal. Yang tidak saya sarankan adalah
> menyalakannya tanpa tahu — makanya disclosure itu ada."

### "Junior jadi nggak belajar dong?"

> "Kekhawatirannya wajar, tapi menurut saya yang berubah levelnya, bukan jumlahnya.
>
> Yang hilang: menghafal sintaks, menulis boilerplate, mencari cara memasang sesuatu.
> Yang justru jadi lebih penting: bisa membaca diff dan tahu mana yang salah, bisa
> merumuskan apa yang harus benar sebelum kodenya ditulis, dan tahu kapan harus tidak
> percaya pada test hijau.
>
> Tiga hal itu dulu butuh bertahun-tahun untuk dilatih karena kesempatannya jarang.
> Sekarang setiap hari."

### "Kenapa Claude Code, bukan Cursor atau yang lain?"

> "Saya tidak akan mengklaim ini yang terbaik — saya cuma yang paling dalam pakai ini.
>
> Yang membuat saya bertahan bukan modelnya, tapi bahwa workflow-nya bisa ditulis sebagai
> skill yang masuk ke repo dan ikut versi kontrol. Pipeline yang kalian lihat hari ini
> ada di repo, bisa direview, bisa diperbaiki. Itu bedanya dengan cara kerja yang cuma
> ada di kepala satu orang.
>
> Sebagian besar yang saya tunjukkan hari ini — konteks, pengukuran, mencatat pelajaran —
> tidak terikat ke alat ini sama sekali."

### "Butuh berapa lama setup sampai bisa seperti ini?"

> "Repo ini butuh beberapa hari, dan sebagian besarnya bukan setup alat — tapi
> membereskan yang sudah ada: dokumentasi yang tidak cocok dengan kode, tooling yang
> rusak, dan membuat dua alat ukur.
>
> Tapi itu bukan biaya yang harus dibayar di depan. Ambil satu fitur, kerjakan serapi
> mungkin, jadikan itu contoh yang ditunjuk di prompt berikutnya. Nilainya menumpuk dari
> situ."

### "Kenapa nggak pakai multi-agent?"

Lihat bagian [Kenapa bukan multi-agent](#kenapa-bukan-multi-agent) di atas.

### Pertanyaan yang tidak punya jawaban

Akan ada. Jawab apa adanya:

> "Saya tidak tahu. Belum saya coba."

Lalu tulis di parking lot. Satu "saya tidak tahu" yang jujur menaikkan kepercayaan pada
semua jawaban lain lebih banyak daripada satu jawaban yang dikarang.

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
