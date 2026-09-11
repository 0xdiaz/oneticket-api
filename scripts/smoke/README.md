# Smoke test

Menjawab pertanyaan yang tidak dijawab test mana pun: **yang barusan naik itu
benar-benar hidup atau tidak.**

```bash
go run ./scripts/smoke                 # default http://localhost:8000
BASE_URL=https://staging go run ./scripts/smoke
```

Sembilan cek, semuanya baca saja, selesai di bawah satu detik:

| # | Cek | Kenapa ada |
|---|---|---|
| 1 | Server menjawab | Kalau ini gagal, sisanya tidak perlu dilaporkan |
| 2 | Database tersambung | `/health` bisa 200 sementara DB-nya putus |
| 3 | Endpoint metrics hidup | - |
| 4 | Migrasi sudah jalan | Tabel ada, bukan cuma koneksi ada |
| 5 | Event ter-seed | Data awal yang dipakai demo |
| 6 | Login berhasil | Jalur tulis paling murah yang bisa dicek |
| 7 | **Route terlindungi menolak tanpa token** | Yang paling penting, lihat di bawah |
| 8 | Route terlindungi menerima dengan token | Pasangan dari nomor 7 |
| 9 | Route tak dikenal menjawab 404 | Fallback masih terpasang |

Nomor 7 yang paling penting dan alasannya ada di kodenya: **guard yang berhenti
menjaga tetap menjawab 200.** Delapan cek lainnya akan lolos semua. Cuma cek yang
mengharapkan penolakan yang bisa melihatnya.

Exit code: `0` sehat, `1` ada cek yang gagal.

`go run` tidak meneruskan exit code ini: ia mencetak `exit status 1` sebagai teks
lalu keluar dengan 1. Kompilasi dulu kalau exit code-nya mau dipakai skrip.

Smoke bukan pengganti test. Ia tidak tahu apakah logikanya benar, ia cuma tahu
apakah yang seharusnya hidup memang hidup. Jalankan setelah deploy, bukan
sebelum commit.
