# Narasi demo

Bukan naskah kata per kata, karena itu terdengar dibacakan. Ini **beat**: apa yang harus
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

## 00:00, Pembukaan (8 menit)

**Tujuan.** Pasang ekspektasi, dan cegah dua salah paham sebelum sempat terbentuk.

Salah paham pertama: orang mengira mereka akan melihat swarm agent otonom. Pakai empat
tingkat itu untuk menempatkan sesi ini:

```
Claude Code single agent       -> mengendalikan 1 agent
Background agent               -> agent otomatis, bisa mereview agent
Human orchestrator multi-agent -> manusia mengendalikan banyak agent
AI orchestrator multi-agent    -> agent mengendalikan agent
```

> "Hari ini kita main di tingkat satu, nyenggol dikit tingkat dua. Bukan karena yang
> di bawah nggak bisa. Tapi karena yang nentuin hasil itu bukan berapa banyak
> agentnya."

Salah paham kedua: orang mengira demo AI selalu diatur. Dahului tuduhannya:

> "Nanti bakal ada bug yang muncul di depan kalian. Itu bukan bug yang saya tanam ya.
> Implementasi yang wajar-wajar aja emang ngasilin bug itu, dan nanti saya tunjukin
> kenapa semua test-nya hijau padahal bugnya ada."

Lalu peta pipeline, dan tunjuk bahwa **semua kotaknya dijalankan hari ini**:

```
Brainstorm -> Planning -> Work + Verifikasi -> Review -> Compound
```

> "Jadi bukan cuplikan. Satu putaran penuh, dari belum ada fiturnya sampai
> pelajarannya kecatat di repo."

---

## 00:08, Kondisi awal repo (8 menit)

**Tujuan.** Ini segmen yang paling sering dilewatkan, dan yang paling menentukan apakah
tesis lo dipercaya. Kalau penonton tidak melihat konteksnya dulu, mereka akan mengira
agentnya sakti.

Tunjukkan, jangan ceritakan:

```bash
wc -l CLAUDE.md docs/00_AI_CRITICAL_RULES.md docs/MODULE_GUIDE.md
ls docs/
```

> "Nah, ini yang dibaca agent sebelum dia nulis satu baris pun. Jadi bukan prompt saya
> yang jago, reponya yang udah bisa ngejelasin dirinya sendiri."

Buka `docs/00_AI_CRITICAL_RULES.md`, scroll bagian *Quick Decision Tree*. Lalu buka satu
slice nyata, `internal/app/services/event_service.go`, dan tunjukkan komentar ini:

```go
// Availability is counted per event, so this issues one query per event on
// top of the list query. Fine at demo scale, and deliberately left that way:
// it is the honest starting point for a later performance pass.
```

> "Perhatiin baris ini. Ini sengaja saya tinggal, dan sengaja dikasih komentar. Inget
> ya, nanti kita balik ke sini."

**Itu menanam puncaknya.** Penonton akan mengenali momennya sendiri nanti, jauh lebih
kuat daripada lo yang mengumumkannya.

Tutup dengan kejujuran yang membangun kepercayaan:

> "Repo ini nggak dari dulu begini. Waktu saya mulai, dokumentasinya ngegambarin
> arsitektur yang nggak ada di kodenya. Lima belas file, sepuluh ribu baris, semuanya
> nyuruh agent bikin di folder yang nggak pernah dibuat.
>
> Jadi yang kalian lihat sekarang itu hasil beresin itu dulu."

---

## 00:16, `/ce-brainstorm` (12 menit) · time box keras

**Tujuan.** Menunjukkan bahwa agent yang bagus **bertanya**, dan pertanyaannya bagus.

Sebelum paste, katakan apa yang akan terjadi:

> "Saya udah tau mau bangun apa. Yang belum saya putusin cuma tiga hal. Coba lihat,
> dia nanya tiga itu, atau malah nawarin fitur yang nggak saya minta."

Paste prompt [1]. Sambil dia berpikir:

> "Prompt-nya panjang bukan karena saya hobi ngetik. Semua yang **udah** diputusin saya
> sebutin di depan, biar dia nggak buang giliran nanyain hal yang jawabannya udah ada.
> Yang saya sisain cuma tiga."

**Ketika pertanyaan muncul, baca jawabannya dari `answers.md`.** Jangan berpikir di
panggung. Itu lambat dan terlihat ragu.

Pertanyaan ketiga adalah yang penting. Jawabannya:

> "Itu keputusan implementasi. Yang saya kunci itu hasilnya: nggak boleh oversell,
> plus test yang buktiin."

Lalu jelaskan kenapa, karena ini pelajaran tersendiri:

> "Kalau tadi saya jawab 'pakai row lock', berarti saya baru aja mutusin hal teknis di
> dokumen requirement. Padahal requirement itu soal apa yang harus bener, bukan
> gimana caranya."

**Kalau lewat 12 menit:** hentikan, `git checkout demo/plan-ready`, bilang apa adanya.

---

## 00:28, `/ce-plan` (15 menit) · time box keras

**Tujuan.** Menunjukkan bahwa "konteks" itu artefak konkret, bukan kata-kata motivasi.

Paste prompt [2]. Sambil menunggu, ini slot terbaik untuk poin TDD:

> "Perhatiin yang saya minta: task nulis test harus **sebelum** task implementasi, dan
> tiap test-nya disebut satu-satu. Kalau nggak diminta gitu, biasanya yang keluar satu
> task namanya 'write tests' di paling bawah. Itu bukan TDD, itu test yang ditempel
> belakangan."

Setelah plan jadi, buka dan tunjuk tabel Requirements Trace:

> "Tiap requirement ada unit yang ngerjain dan test yang buktiin. Kalau ada baris yang
> kosong unitnya, berarti plan-nya belum kelar. Jadi tabel ini bukan dokumentasi, ini
> alat cek."

---

## 00:43, `/ce-work` (22 menit) · jendela Q&A terbesar

**Tujuan.** Ini bagian yang paling sedikit butuh narasi dan paling banyak butuh Q&A.

Paste prompt [3]. Katakan sekali di depan:

> "Oke, sekarang dia kerja. Ini makan dua puluh menitan, jadi mending kita pakai
> waktunya, silakan tanya apa aja."

**Kalau tidak ada yang bertanya**, pancing dengan salah satu dari empat ini. Yang
terakhir paling ampuh:

1. "Ini bedanya apa sama autocomplete di IDE?"
2. "Kalau agentnya salah dan nggak ketahuan gimana?"
3. "Ini habis berapa token sekali jalan?"
4. **"Tim saya nggak punya dokumentasi sebagus ini. Mulai dari mana?"**

Jawaban untuk yang keempat, karena ini pertanyaan yang paling sering ditanya dan paling
berguna dijawab dengan baik:

> "Jangan mulai dari nulis dokumentasi. Mulai dari satu slice yang bener aja, satu
> fitur, ditulis serapi mungkin, test-nya jujur. Terus tinggal tunjuk itu di prompt:
> 'ikutin pola yang di sini'.
>
> Satu contoh nyata itu jauh lebih kuat daripada lima ribu baris standar. Dokumentasinya
> nyusul belakangan, bukan duluan."

**Minta satu orang memberi kode isyarat waktu agent berhenti.** Lo tidak akan sadar
sendiri karena sedang bicara.

---

## 01:05, Puncak (15 menit)

Dua alat, dua pelajaran. Ini bagian yang menentukan apakah sesi lo diingat.

### 4a, Bug yang *dicegah*

```bash
./scripts/loadtest/reset.sh
go run ./scripts/loadtest -n 300 -c 80
```

Layar: `AMAN, terjual 100 dari 100 tiket`.

Di sini ada godaan untuk buru-buru lewat. Jangan. **Ini poin pertama lo, dan tidak
terlihat seperti poin kalau tidak dijelaskan:**

> "Ini race condition klasik ya. Seratus tiket, tiga ratus orang mencet tombol
> barengan. Implementasi naif bakal jual tiga ratus. Dan barusan kalian lihat itu
> nggak kejadian.
>
> Bukan karena modelnya pinter. Tapi karena plan-nya nyebut transaksi, nyebut unique
> constraint, dan yang paling nentuin: test konkurennya udah ada di plan, kelihatan
> sama dia sebelum dia nulis implementasinya. Dia udah tau bakal diuji kayak gimana."

Tunjukkan barisnya:

```bash
grep -A3 'SKIP LOCKED' internal/domain/repositories/order_repo.go
```

> "Jadi konteks yang bagus itu nggak bikin agent-nya tambah pinter. Dia bikin jawaban
> yang bener jadi jawaban yang paling kelihatan."

### 4b, Bug yang *lolos semua test*

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

> "Dua ratus dua query buat satu request. Dan semua test-nya hijau.
>
> Kenapa? Karena hasilnya **bener**. Cuma mahal. Nggak ada assertion yang gagal, nggak
> ada error di log, nggak ada yang merah. Test itu nguji kebenaran, bukan biaya.
>
> Nah ini kelas bug yang beda dari yang tadi. Yang tadi kecegah sama konteks yang bagus.
> Yang ini nggak kecegah apa-apa, dan nggak bakal ketahuan sampai ada yang ngukur."

Paste prompt `/ce-debug`. Sambil dia bekerja:

> "Perhatiin, saya nggak nyuruh dia benerin. Saya nyuruh dia reproduce dulu, bikin
> test yang gagal karena biayanya, bukan karena hasilnya. Terus sebutin opsi
> perbaikannya beserta trade-off-nya, baru milih.
>
> Bagian nimbang-nimbang itu yang biasanya nggak kelihatan."

Setelah fix:

```bash
go run ./scripts/nplusone
```

`AMAN, query tetap 2 meski event naik dari 10 ke 200`.

> "Dua ratus dua jadi dua."

Biarkan angka itu menggantung sebentar. Jangan langsung lanjut.

---

## 01:20, `/ce-code-review` (10 menit) · jendela Q&A

**Tujuan.** Menunjukkan review yang berjalan paralel dan lintas model.

> "Yang jalan sekarang beberapa persona sekaligus, plus satu peer di model lain.
> Alasannya sederhana: model yang nulis kodenya itu pembaca paling jelek buat kode itu.
> Dia udah yakin kodenya bener, kan dia baru aja ngeyakinin dirinya sendiri."

Kandidat pertama yang dipotong kalau waktu mepet. Kalau dipotong, cukup katakan apa yang
biasanya dia temukan.

---

## 01:30, `/ce-compound` (7 menit)

**Tujuan.** Menunjukkan mekanisme yang jadi nama seluruh pendekatan ini.

Setelah selesai:

```bash
ls docs/solutions/
git log --oneline -5
```

> "Pelajaran yang tadi sekarang udah ada di repo, dan udah ke-commit. Bukan di kepala
> saya, bukan di catatan pribadi, bukan di Slack yang dua minggu lagi ilang."

> "Ini bagian yang biasanya dilewat. Padahal justru ini satu-satunya yang bikin putaran
> berikutnya lebih murah dari yang barusan."

---

## 01:37, Story kedua (13 menit) · jendela Q&A

**Tujuan.** Membuktikan klaimnya. Jangan potong segmen ini.

Paste prompt [7]. Sebelum agent mulai, buat prediksi yang bisa jatuh:

> "Di prompt ini saya sama sekali nggak nyebut locking. Kalau nanti dia pakai pola yang
> sama kayak story pertama, itu karena dia baca `docs/solutions/`, bukan karena saya
> suruh."

Membuat prediksi di depan itu berisiko, dan justru itu yang membuatnya bernilai. Kalau
meleset, katakan meleset, penonton akan lebih percaya semua yang lain.

Kalau tepat, tunjukkan berdampingan: entri `docs/solutions/` di kiri, kode baru di kanan.

> "Nah, ini yang dimaksud compound. Yang numpuk bukan kodenya, tapi pelajarannya."

---

## 01:50, Penutup (8 menit)

Empat hal, jangan lebih:

1. **Konteks > prompt.** Yang membuat outputnya konsisten adalah repo yang menjelaskan
   dirinya, bukan prompt yang panjang.
2. **Agent harus bisa mengukur.** Test membuat dia mandiri. Tapi test menguji kebenaran,
   untuk biaya dan konkurensi, butuh alat ukur sendiri.
3. **Learning di-commit.** Kesalahan yang tidak dicatat akan terjadi lagi di story
   berikutnya.
4. **Review tetap manusia.** Yang berubah levelnya: dari mengecek sintaks jadi mengecek
   keputusan.

Lalu tutup dengan yang paling jujur, karena ini yang membedakan sesi lo dari konten
"AI bikin saya 10x lebih cepat":

> "Yang kalian lihat hari ini bukan agent-nya yang hebat.
>
> Sebagian besar kerja saya minggu ini bukan nulis prompt. Tapi beresin dokumentasi biar
> cocok sama kodenya, benerin tooling yang rusak, dan bikin dua alat ukur. Habis itu
> baru agent-nya kepake.
>
> Dan urutannya penting. Kalau repo kalian belum bisa ngejelasin dirinya sendiri,
> nambahin agent itu cuma bikin berantakannya lebih cepet."

---

## Kenapa bukan multi-agent

Pertanyaan ini hampir pasti datang, dan jawabannya memperkuat tesis lo, jadi jangan
defensif. Sebut sendiri di pembukaan, satu kalimat setelah empat tingkat agent:

> "Tingkat tiga sama empat itu ada, di Claude Code namanya agent teams. Hari ini sengaja
> nggak dipakai, nanti saya jelasin kenapa. Bukan karena belum sempet nyoba ya."

Lalu saat ditanya, tiga alasan. Yang ketiga yang paling penting.

**Satu, bentuk pekerjaannya tidak cocok.** Dokumentasinya sendiri menulis: *"For
sequential tasks, same-file edits, or work with many dependencies, a single session or
subagents are more effective."* Implementasi di demo ini persis itu: unit yang saling
bergantung, file yang sama. Teams unggul untuk riset paralel, review, dan debat
hipotesis, bukan implementasi berurutan.

**Dua, menyalakannya mengubah hal lain.** *"A subagent that Claude names launches as a
teammate, so teams can form even when you didn't ask for one."* `ce-code-review` memang
menyebar subagent bernama; begitu teams menyala, roster itu berubah jadi teammate. Fitur
eksperimental yang diam-diam mengubah bagian pipeline yang sudah diuji bukan hal yang
dinyalakan seminggu sebelum tampil.

**Tiga, lima Claude punya blind spot yang sama.** Ini yang benar-benar penting:

> "Teammate itu kan instance Claude Code juga. Model yang sama, salahnya dengan cara
> yang sama, cuma lebih banyak. Lima agent debat-debatan ya tetep lima agent dengan
> asumsi yang mirip-mirip.
>
> Yang saya butuh itu satu yang **salahnya beda**. Nambah agent nggak nyelesain itu.
> Makanya review di sini lewat model lain, Codex."

Paralelisme menambah kecepatan. Model yang berbeda menambah **sudut pandang**. Untuk
review, yang kedua jauh lebih berharga.

---

## Bank jawaban Q&A

Jawaban yang sudah disiapkan untuk pertanyaan yang kemungkinan besar datang. Ambil
intinya, jangan hafalkan kalimatnya.

### "Bedanya apa sama Copilot atau autocomplete di IDE?"

> "Autocomplete itu nerusin kalimat yang lagi saya ketik. Yang tadi kalian lihat dia
> baca plan, nulis migrasi, nulis test, lihat test-nya gagal, baru nulis
> implementasinya, terus ngukur hasilnya sendiri.
>
> Bedanya sebenernya bukan di modelnya. Autocomplete nggak punya konsep 'selesai'.
> Agent punya. Ada definition of done, ada gate verifikasi, dan dia tau dia belum
> kelar sampai gate-nya hijau."

### "Berapa biayanya sekali jalan?"

Jangan mengarang angka. Yang jujur:

> "Saya nggak punya angka pastinya, belum saya ukur per-run. Tapi bentuk biayanya kira-
> kira gini: review itu yang paling mahal, karena dia nyebar beberapa persona plus satu
> model lain. Implementasi malah jauh lebih murah dari yang orang kira, soalnya
> sebagian besar token-nya kepake buat baca, bukan nulis.
>
> Tapi jujur, biaya yang lebih penting itu bukan token. Yang paling makan waktu minggu
> ini justru beresin reponya sampai agent-nya kepake. Jalanin agent-nya sendiri cepet."

### "Kalau agentnya salah dan nggak ketahuan gimana?"

Ini pertanyaan terbaik yang bisa datang. Jawab dengan contoh nyata dari repo ini:

> "Itu kejadian di repo ini, dan bukan kasus kecil.
>
> Endpoint forgot-password itu ngembaliin token reset password di body respons. Jadi
> siapa pun yang tau email korban bisa ambil akunnya cuma dengan dua request. Dan test
> suite-nya **hijau**.
>
> Bukan cuma hijau. Ada test yang secara eksplisit mastiin token itu dikembaliin. Jadi
> siapa pun yang benerin bakal lihat test-nya merah, terus ngira dia sendiri yang
> salah.
>
> Jadi jawabannya: test hijau itu bukti yang lebih lemah dari yang kita kira. Makanya
> di sesi ini saya nggak cuma jalanin test, saya ngukur. Load test buat konkurensi,
> query probe buat biaya. Dua-duanya nangkep hal yang nggak ketangkep test."

### "Ini bisa dipakai di codebase legacy yang berantakan?"

> "Bisa, tapi urutannya kebalik dari yang orang kira.
>
> Repo ini waktu saya mulai punya lima belas dokumen yang ngegambarin arsitektur yang
> nggak ada di kodenya. Dokumen yang paling pertama dibaca agent malah nyuruh bangun di
> folder yang nggak pernah dibuat. Jadi agent yang nurut ke dokumentasi itu bakal
> ngasilin kode yang yatim piatu di tengah codebase.
>
> Nah buat legacy: jangan mulai dengan nyuruh agent ngerjain fitur. Mulai dengan nyuruh
> dia **baca satu area, terus laporin apa yang sebenernya ada di situ**, baru
> dokumennya dibenerin. Itu justru kerjaan yang cocok banget buat agent, dan hasilnya
> jadi modal buat semua kerjaan setelahnya."

### "Apa yang agent tetap tidak bisa?"

> "Dia nggak bisa tau apa yang seharusnya bener kalau nggak ada yang ngasih tau.
>
> Contoh dari hari ini. Ada test yang saya minta dia tulis buat event dengan nol tiket.
> Databasenya nolak, karena ada CHECK constraint yang ngelarang event tanpa inventaris.
> Jadi test-nya yang salah, bukan kodenya. Dan yang ngoreksi bukan agent, bukan saya,
> tapi database yang megang aturannya.
>
> Itu polanya. Agent jago banget ngikutin aturan yang tertulis, tapi buta sama aturan
> yang cuma ada di kepala orang. Makanya aturan yang penting harus ditaruh di tempat
> yang gagalnya berisik, constraint database, compile error, test. Bukan kalimat di
> dokumen."

### "Kode kita dikirim ke mana? Aman?"

Jawab lurus, jangan berkelit:

> "Buat review, iya. Ada satu pass yang dikirim ke model lain, dan skill-nya emang
> ngumumin itu sebelum jalan. Repo yang saya pakai hari ini publik, jadi nggak masalah.
>
> Kalau buat repo kerjaan, itu keputusan kebijakan, bukan teknis. Pass-nya bisa
> dimatiin, dan ada fallback reviewer lokal. Yang nggak saya saranin itu nyalain tanpa
> tau, makanya disclosure-nya ada."

### "Junior jadi nggak belajar dong?"

> "Kekhawatirannya wajar. Tapi menurut saya yang berubah itu levelnya, bukan jumlahnya.
>
> Yang ilang: ngapalin sintaks, nulis boilerplate, nyari cara masang sesuatu. Yang
> justru jadi lebih penting: bisa baca diff dan tau mana yang salah, bisa ngerumusin apa
> yang harus bener sebelum kodenya ditulis, dan tau kapan harus nggak percaya sama test
> hijau.
>
> Tiga hal itu dulu butuh tahunan buat dilatih, soalnya kesempatannya jarang. Sekarang
> tiap hari."

### "Kenapa Claude Code, bukan Cursor atau yang lain?"

> "Saya nggak bakal klaim ini yang paling bagus, saya cuma paling dalem makenya yang
> ini.
>
> Yang bikin saya bertahan itu workflow-nya bisa ditulis jadi skill yang masuk repo dan
> ikut version control. Bukan soal modelnya. Pipeline yang kalian lihat hari ini ada di repo,
> bisa direview, bisa diperbaiki. Beda sama cara kerja yang cuma ada di kepala satu
> orang.
>
> Dan sebagian besar yang saya tunjukin hari ini, konteks, ngukur, nyatet pelajaran,
> nggak keiket sama alat ini sama sekali."

### "Butuh berapa lama setup sampai bisa seperti ini?"

> "Repo ini butuh beberapa hari, dan sebagian besarnya habis buat beresin yang udah ada.
> Dokumentasi yang nggak cocok sama kode, tooling yang rusak, sama bikin dua alat ukur.
> Setup alatnya sendiri sebentar.
>
> Tapi itu bukan biaya yang harus dibayar di depan. Ambil satu fitur, kerjain serapi
> mungkin, terus jadiin itu contoh yang ditunjuk di prompt berikutnya. Nilainya numpuk
> dari situ."

### "Kenapa nggak pakai multi-agent?"

Lihat bagian [Kenapa bukan multi-agent](#kenapa-bukan-multi-agent) di atas.

### Pertanyaan yang tidak punya jawaban

Akan ada. Jawab apa adanya:

> "Saya nggak tau. Belum pernah saya coba."

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
