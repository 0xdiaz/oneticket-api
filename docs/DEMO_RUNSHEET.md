# Runsheet — Sharing Session: Agentic AI-Driven Development

**Format:** meetup publik, 120 menit, Q&A menyatu ke dalam sesi
**Kendaraan demo:** Flash Sale Ticketing API di atas `gin-boilerplate`
**Pipeline:** Compound Engineering (BMad tidak dipakai — lihat "Kenapa CE saja")
**Yang dijual:** bukan produknya, tapi *agent yang dikasih konteks bagus bisa dikasih kerjaan beneran*

---

## Thesis

> Agentic coding gagal bukan karena modelnya bodoh, tapi karena konteksnya kosong.
> Repo ini punya dokumentasi yang sudah disamakan dengan kodenya, aturan AI yang
> eksplisit, dan `docs/solutions/` yang tumbuh tiap story. Itu sebabnya agent-nya nurut.

Semua bagian sesi harus menopang kalimat itu. Kalau ada segmen yang nggak nyambung ke
situ, potong.

---

## Dua jenis nunggu

Ini yang menentukan seluruh struktur di bawah.

| Jenis | Contoh | Bisa diisi Q&A? |
|---|---|---|
| **Agent kerja sendiri** | `ce-work`, `ce-code-review`, `ce-compound` | ✅ Ya — ini jendela Q&A-nya |
| **Agent nanya balik** | `ce-brainstorm`, `ce-plan` | ❌ Tidak — lo harus jawab agent, bukan audiens |

Karena itu planning **tetap di-pre-bake meski waktu cukup**. Menjalankan `ce-plan` live
bukan menghemat apa-apa: audiens cuma nonton lo ngobrol sama terminal.

---

## Timing

| Waktu | Durasi | Segmen | Q&A? |
|---|---|---|---|
| 00:00 | 10' | **Pembukaan + thesis** | — |
| 00:10 | 10' | **Kondisi awal repo** — AI rules, doc yang sudah disamakan, test yang ada | — |
| 00:20 | 8' | **Artefak planning** — hasil `ce-brainstorm` + `ce-plan` (pre-baked) | — |
| 00:28 | 25' | 🎬 **`/ce-work docs/plans/checkout.md`** — agent implement checkout | ✅ **jendela besar** |
| 00:53 | 15' | 🎬 **Puncak** — load test → `OVERSELL 300/100` → `/ce-debug` → fix → `100/100` | sedikit |
| 01:08 | 12' | 🎬 **`/ce-code-review`** — fan-out persona paralel + peer cross-model | ✅ **jendela** |
| 01:20 | 8' | 🎬 **`/ce-compound`** — learning masuk `docs/solutions/` | — |
| 01:28 | 15' | 🎬 **Story kedua (refund)** — agent pakai pola locking tanpa disuruh | ✅ **jendela** |
| 01:43 | 12' | **Takeaway + Q&A terbuka** | ✅ |
| 01:55 | 5' | **Buffer** | — |

**Total 115 menit + 5 buffer.** Kalau molor, yang dipotong `ce-code-review` — **bukan**
segmen debug, dan **bukan** story kedua (itu thesis-nya).

---

## Kenapa CE saja, tanpa BMad

BMad itu upacara sprint: story file, `sprint-status.yaml`, epic. Berguna, tapi bukan yang
mau dibuktikan hari ini. Compound Engineering itu mekanisme *compounding* — dan
`docs/solutions/` persis mekanisme yang jadi tesis sesi ini.

Satu hal yang **tidak boleh** dijalankan di panggung: `/bmad-sprint-run`. Definisinya
sendiri bilang *"execute ALL stories in ALL epics… no pauses, no confirmation prompts, no
stopping until every story reaches done or blocked"*, dengan retry budget 3 per story.
Itu tool buat ditinggal semalaman. Sekali diketik, kendali atas jam dinding hilang.

---

## Kenapa arc-nya begini

```
ce-work jalan mulus     → "oh, bisa nulis kode"          (belum meyakinkan)
Oversell 300/100        → "nah, tetep bisa salah kan"    (skeptis menang)
ce-debug diagnosa+fix   → "...oke ini beda"              (skeptis goyah)
ce-compound catat       → "jadi nggak ngulang"           (paham mekanismenya)
Story 2 pakai pola      → "ini yang namanya compounding" (takeaway kebawa pulang)
```

Skeptis di ruangan **harus dikasih menang dulu** di menit 53. Kalau semuanya mulus dari
awal, mereka pulang dengan pikiran "ah, demo-nya diatur".

---

## Aturan desain demo: semua jalan keluar harus menang

Agent itu non-deterministik. Jangan pernah bikin demo yang cuma sukses kalau agent milih
satu solusi tertentu.

Untuk oversell, agent bisa fix pakai:
- `SELECT ... FOR UPDATE` (row lock)
- `SELECT ... FOR UPDATE SKIP LOCKED`
- Optimistic locking + retry
- `UPDATE ... WHERE status = 'available'` atomik
- Unique constraint di tabel order

**Semuanya jawaban benar.** Siapkan komentar singkat untuk masing-masing, jadi apa pun
yang keluar lo bisa bilang "nah, dia pilih X — trade-off-nya Y". Itu justru lebih
meyakinkan daripada hasil yang seragam.

---

## Mengelola Q&A yang menyatu

Jendela Q&A adalah bagian paling rapuh dari format ini. Tiga hal yang bikin gagal:

**1. Nggak ada yang nanya.** Siapkan 4 pertanyaan pancingan yang lo tanya ke diri sendiri:
- "Ini bedanya apa sama autocomplete di IDE?"
- "Kalau agent-nya salah dan nggak ketahuan gimana?"
- "Ini habis berapa token / berapa duit sekali jalan?"
- "Tim gue nggak punya dokumentasi sebagus ini, mulai dari mana?"

**2. Q&A kelewat panjang, agent udah selesai dari tadi.** Lo nggak akan sadar karena lagi
ngomong. Taruh terminal di layar yang kelihatan ekor mata lo, atau minta satu orang kasih
kode isyarat waktu agent berhenti.

**3. Ada pertanyaan yang butuh layar.** Itu ngerebut layar dari agent yang lagi jalan.
Siapkan **parking lot** — tulis di papan, jawab di segmen 01:43.

---

## Pre-flight checklist (H-1)

- [ ] `go mod download` selesai — **jangan pernah nunggu download di depan orang**
- [ ] Postgres jalan: `docker compose --env-file .env -f .docker/docker-compose-dev.yml up -d postgres_db`
      (`--env-file` wajib — compose ada di `.docker/`, tanpa itu interpolasi `${MASTER_DB_*}` kosong)
- [ ] `go run main.go` sekali: migration keapply, event "Flash Sale Demo" + 100 tiket ke-seed
- [ ] `curl localhost:8000/api/v1/events` sudah balas `available_tickets: 100`
- [ ] `./scripts/loadtest/reset.sh` jalan, lalu `go run ./scripts/loadtest -n 300 -c 80`
      balas `BELUM ADA YANG TERJUAL` (404) — itu kondisi awal yang benar
- [ ] `RATE_LIMIT_RPS=1000` di `.env` — di 100, rate limiter nolak duluan dan
      **oversell-nya nggak akan pernah muncul**
- [ ] Artefak `ce-brainstorm` + `ce-plan` sudah ada di `docs/plans/`
- [ ] Semua prompt ada di file teks, tinggal paste — ngetik prompt live itu dead air
- [ ] Branch parachute `demo/final` yang sudah jadi dan sudah diverifikasi
- [ ] **Dry run persis sekali** dengan prompt yang sama, catat menitnya per segmen
- [ ] Video cadangan full run direkam semalam sebelumnya
- [ ] Font gede, notifikasi mati, layar dibagi: agent kiri / output test kanan

## Yang perlu diwaspadai dari skill CE

| Skill | Perilaku yang bisa bikin kaget |
|---|---|
| `ce-work` | Punya *shipping tail* — bisa commit/push/PR sendiri. Pakai `mode:return-to-caller` kalau nggak mau |
| `ce-code-review` | Mengirim kode ke peer model lain (*cross-model egress*). Repo publik jadi aman, tapi ini segmen paling lambat |
| `ce-compound` | Ikut commit `docs/solutions/`. Justru bagus: `git log` jadi bukti di layar |

## Pas jalan

- Pas agent mikir — **itu slot lo**, bukan diem. Selain Q&A, siapkan 3 poin isian:
  1. Kenapa `docs/00_AI_CRITICAL_RULES.md` ada dan apa isinya
  2. Kenapa plan file lebih bagus daripada prompt panjang
  3. Kenapa review pakai model lain itu penting
- Kalau agent nyasar: **maksimal 3 attempt**. Lewat itu, `git checkout demo/final`,
  jelasin apa yang tadi salah, lanjut. Audiens lebih respek jujur daripada demo dipaksa mulus
- Jangan sembunyiin error. Momen agent salah lalu benerin sendiri itu **konten paling mahal**

---

## Kontingensi

| Kalau… | Lakukan |
|---|---|
| Internet mati | Video cadangan, narasikan langsung |
| Agent stuck > 3 attempt | `git checkout demo/final`, bahas kenapa gagal |
| Load test nggak nunjukkan oversell | Naikkan `-n` dan `-c`, cek `RATE_LIMIT_RPS`, atau tambah `time.Sleep` di service (dan **bilang** ke audiens kalau lo lakuin itu) |
| `ce-code-review` kelamaan | Potong, lanjut ke `ce-compound` |
| Kehabisan waktu | Potong `ce-code-review` dulu, baru buffer. Jangan potong story kedua |
| Agent nge-fix duluan sebelum load test | Bagus — tunjukkan test-nya, bahas kenapa dia antisipasi |
| Nggak ada yang nanya di jendela Q&A | Pakai 4 pertanyaan pancingan di atas |

---

## Takeaway yang harus kebawa pulang

1. **Konteks > prompt.** Dokumentasi dan aturan di repo yang bikin outputnya konsisten
2. **Agent harus bisa mengukur.** Test dan load test yang bikin dia bisa mandiri, bukan modelnya
3. **Learning di-commit.** `docs/solutions/` bikin kesalahan cuma terjadi sekali
4. **Review tetap manusia.** Yang berubah levelnya — dari cek sintaks jadi cek keputusan
