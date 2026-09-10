# Runsheet — Sharing Session: Agentic AI-Driven Development

**Format:** meetup publik, 75 menit (buffer sampai 90)
**Kendaraan demo:** Flash Sale Ticketing API di atas `gin-boilerplate`
**Yang dijual:** bukan produknya, tapi *agent yang dikasih konteks bagus bisa dikasih kerjaan beneran*

---

## Thesis

> Agentic coding gagal bukan karena modelnya bodoh, tapi karena konteksnya kosong.
> Repo ini punya 5000+ baris dokumentasi dan AI rules. Itu sebabnya agent-nya nurut.

Semua bagian sesi harus menopang kalimat itu. Kalau ada segmen yang nggak nyambung ke situ, potong.

---

## Timing

| Waktu | Durasi | Segmen | Tujuan |
|---|---|---|---|
| 00:00 | 8' | **Pembukaan + thesis** | Pasang ekspektasi: ini bukan demo sulap |
| 00:08 | 7' | **Kondisi awal repo** | Tunjukkan `docs/00_AI_CRITICAL_RULES.md`, struktur clean arch, test yang udah ada |
| 00:15 | 5' | **Planning** | PRD + story udah jadi (pre-bake). Tunjukkan *cara* bikinnya, jangan bikin live |
| 00:20 | 18' | **🎬 Demo 1 — Story: checkout tiket** | Agent implement. Test hijau. Endpoint jalan |
| 00:38 | 10' | **🎬 Puncak — oversell** | Load test → `137/100` → `/ce-debug` → fix → `100/100` |
| 00:48 | 7' | **`/ce-compound`** | Learning masuk `docs/solutions/`. Tunjukkan isinya |
| 00:55 | 10' | **🎬 Demo 2 — Story: refund** | Agent pakai pola locking dari langkah sebelumnya **tanpa disuruh** |
| 01:05 | 10' | **Takeaway + Q&A** | |

**Total 75 menit.** Kalau molor, yang dipotong: Demo 2 (bukan puncaknya). Kalau kecepetan, panjangin Q&A.

---

## Kenapa arc-nya begini

```
Story 1 mulus        → "oh, bisa nulis kode"          (belum meyakinkan)
Oversell muncul      → "nah, tetep bisa salah kan"    (audiens skeptis menang)
Agent diagnosa+fix   → "...oke ini beda"              (skeptis mulai goyah)
Learning dicatat     → "jadi nggak ngulang"           (paham mekanismenya)
Story 2 pakai pola   → "ini yang namanya compounding" (takeaway kebawa pulang)
```

Skeptis di ruangan **harus dikasih menang dulu** di menit 38. Kalau semuanya mulus dari awal, mereka pulang dengan pikiran "ah, demo-nya diatur".

---

## Aturan desain demo: semua jalan keluar harus menang

Agent itu non-deterministik. Jangan pernah bikin demo yang cuma sukses kalau agent milih satu solusi tertentu.

Untuk oversell, agent bisa fix pakai:
- `SELECT ... FOR UPDATE` (row lock)
- Optimistic locking + retry
- `UPDATE ... WHERE stock > 0` atomik
- CHECK constraint di DB

**Keempatnya jawaban benar.** Siapin komentar singkat untuk masing-masing, jadi apa pun yang keluar lo bisa bilang "nah, dia pilih X — trade-off-nya Y". Itu justru lebih meyakinkan daripada hasil yang seragam.

---

## Pre-flight checklist (H-1)

- [ ] `go mod download` selesai — **jangan pernah nunggu download di depan orang**
- [ ] Postgres jalan: `docker compose --env-file .env -f .docker/docker-compose-dev.yml up -d postgres_db`
      (`--env-file` wajib — compose ada di `.docker/`, tanpa itu interpolasi `${MASTER_DB_*}` kosong)
- [ ] `go run main.go` sekali: migration keapply, event "Flash Sale Demo" + 100 tiket ke-seed
- [ ] `curl localhost:8000/api/v1/events` sudah balas `available_tickets: 100`
- [ ] `go build ./... && go test ./tests/unit/...` hijau
- [ ] Tool load test kepasang & udah dicoba (`hey`, `vegeta`, atau `k6`)
- [ ] Semua prompt ada di `docs/demo/prompts.md`, tinggal paste — **jangan ngetik prompt live**
- [ ] Branch parachute `demo/final` udah ada dan udah diverifikasi jalan
- [ ] **Dry run lengkap sekali**, catat menit tiap segmen
- [ ] Video cadangan full run direkam
- [ ] Font terminal gede, notifikasi mati, layar dibagi: agent kiri / output kanan

## Pas jalan

- Agent mikir 60 detik = **slot lo ngomong**, bukan diem. Siapin 3 poin isian:
  1. Kenapa `docs/00_AI_CRITICAL_RULES.md` ada dan apa isinya
  2. Kenapa story file lebih bagus daripada prompt panjang
  3. Kenapa review pakai model lain (cross-model) itu penting
- Kalau agent nyasar: **maksimal 3 attempt**. Lewat itu, `git checkout demo/final`, jelasin apa yang tadi salah, lanjut. Audiens lebih respek jujur daripada demo dipaksa mulus
- Jangan sembunyiin error. Momen agent salah lalu benerin sendiri itu **konten paling mahal** di sesi ini

---

## Kontingensi

| Kalau… | Lakukan |
|---|---|
| Internet mati | Video cadangan, narasikan langsung |
| Agent stuck > 3 attempt | `git checkout demo/final`, bahas kenapa gagal |
| Load test nggak nunjukkan oversell | Naikkan concurrency, atau tambah `time.Sleep` di service (dan **bilang** ke audiens kalau lo lakuin itu) |
| Kehabisan waktu | Potong Demo 2, langsung ke takeaway |
| Agent nge-fix duluan sebelum load test | Bagus — tunjukkan test-nya, bahas kenapa dia antisipasi |

---

## Takeaway yang harus kebawa pulang

1. **Konteks > prompt.** Dokumentasi dan aturan di repo yang bikin outputnya konsisten
2. **Agent harus bisa mengukur.** Test dan load test itu yang bikin dia bisa mandiri, bukan modelnya
3. **Learning di-commit.** `docs/solutions/` bikin kesalahan cuma terjadi sekali
4. **Review tetap manusia.** Yang berubah adalah levelnya — dari cek sintaks jadi cek keputusan
