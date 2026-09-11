# Lembar jawaban, hasil dry run

Pertanyaan yang benar-benar muncul saat `/ce-brainstorm` dan `/ce-plan` dijalankan
pada 2026-09-11, beserta jawabannya. **Bacakan, jangan pikirkan.** Memikirkan
keputusan produk sambil bicara adalah cara tercepat melewati time box.

Parasut: `git checkout demo/plan-ready`. Branch itu berisi kedua artefak versi jadi,
`docs/brainstorm/2026-09-11-checkout-requirements.md` (apa yang dibangun) dan
`docs/plans/2026-09-11-001-feat-checkout-plan.md` (bagaimana membangunnya). Keduanya
sengaja tidak ada di `main`: demo yang memproduksinya live.

---

## Yang ditemukan dry run, sebelum daftar jawaban

**1. "Tanyakan sekaligus dalam satu giliran" tidak akan dituruti.**
Interaction Rule 1 di `ce-brainstorm`: *"Ask one question at a time, one question
per turn, even when sub-questions feel related."* Aturan skill menang atas prompt.
Tiga keputusan = **tiga ronde blocking**, bukan satu.

**2. `ce-plan` punya gate konfirmasi sendiri.**
Phase 5.1.5 (scoping synthesis untuk run yang bersumber dari brainstorm) wajib
menunggu konfirmasi untuk depth Standard, bahkan ketika tidak ada call-out.
Ditambah menu handoff wajib di Phase 5.4 setelah plan ditulis.

**Perkiraan total interaksi blocking: 5–6 ronde.** Time box 12 + 15 menit realistis
hanya kalau setiap jawaban dibacakan, bukan dipikirkan.

---

## `/ce-brainstorm`

**Q1. Kepemilikan tiket setelah terjual dimodelkan bagaimana?**
> Tabel `orders` terpisah. Harga saat beli butuh tempat sendiri, dan itu tidak ada
> di baris tiket.

**Q2. Perilaku waktu tiket habis?**
> `409 Conflict`. Konsisten dengan pola error yang sudah ada, dan `utils.Conflict`
> sudah tersedia.

**Q3. Cara mengambil tiket yang tersedia?**
> **Jangan dijawab dengan mekanisme.** Jawab: "itu keputusan implementasi. Yang
> saya kunci itu hasilnya: tidak boleh oversell, plus test yang membuktikannya."
>
> Ini penting: begitu jawaban ini menyebut nama mekanisme, `ce-work` akan
> mengimplementasikannya dengan benar sejak awal dan oversell tidak akan pernah
> terjadi. Puncak demo hilang.

---

## `/ce-plan`

**Q4. `orders` dikasih kolom `status` sekarang, atau nanti saat refund?**
> Sekarang, dengan nilai `paid`. Refund tinggal mengubah nilainya.

**Q5. Test konkuren masuk suite default atau dipisah?**
> Masuk suite default. 50 goroutine memperebutkan 20 tiket, cukup memicu race,
> cukup ringan untuk budget 3 menit. Jangan dipisah build tag: regresi race yang
> hanya ketahuan kalau seseorang ingat menjalankannya bukan jaring pengaman.

**Q6. Menu handoff setelah plan ditulis.**
> Pilih opsi `Start ce-work`, atau tutup menunya dan jalankan `/ce-work` manual
> sesuai `prompts.md` (`mode:return-to-caller`).

---

## Kalau agent menanyakan hal di luar daftar ini

Kemungkinan besar itu scope yang sudah dikeluarkan. Jawab pendek:

> "Di luar ruang lingkup story ini. Refund story berikutnya."

Yang sudah dikeluarkan: refund, waiting list, seat map, pembayaran, hold
sementara, beli lebih dari satu tiket sekali jalan, pemilihan kursi.
