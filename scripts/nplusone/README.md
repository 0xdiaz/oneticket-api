# Query count probe

Mengukur berapa query yang dihabiskan satu panggilan `GET /api/v1/events`,
seiring bertambahnya jumlah event. Query yang ikut tumbuh bersama jumlah baris
adalah tanda tangan N+1.

## Kenapa alat terpisah, bukan test

N+1 **lolos seluruh test suite.** Hasilnya benar, cuma mahal. Tidak ada
assertion yang gagal, tidak ada error di log, tidak ada yang merah. Itulah yang
membuatnya berbahaya — dan kenapa butuh alat yang mengukur biaya, bukan
kebenaran.

Ini bug dengan pelajaran berbeda dari race condition: race ketahuan kalau
diuji konkurensi, N+1 tidak ketahuan sampai ada yang mengukur.

## Usage

```bash
# Postgres jalan
docker compose --env-file .env -f .docker/docker-compose-dev.yml up -d postgres_db

go run ./scripts/nplusone
go run ./scripts/nplusone -sizes 10,100,500
```

Alat ini menanam eventnya sendiri dengan awalan nama `nplusone-probe-` dan
menghapusnya lagi setelah selesai, jadi aman dijalankan terhadap database
development tanpa mengganggu data demo.

## Membaca hasilnya

Sebelum diperbaiki:

```
   EVENT       QUERY        DURASI  QUERY/EVENT
      10          12       7.402ms  1.20
     200         202      16.835ms  1.01

N+1 — query ikut tumbuh bersama jumlah baris
```

Sesudah diperbaiki:

```
   EVENT       QUERY        DURASI  QUERY/EVENT
      10           2       3.541ms  0.20
     200           2       4.114ms  0.01

AMAN — query tetap 2 meski event naik dari 10 ke 200
```

Kolom `QUERY/EVENT` yang mendekati 1,0 adalah tanda tangannya. Yang penting
bukan angka absolutnya, tapi **kemiringannya**: satu request seharusnya
berharga tetap, berapa pun barisnya.

Kedua keadaan sudah diverifikasi sebelum alat ini di-commit.

## Exit code

| Kode | Arti |
|---|---|
| 0 | Query tetap — tidak ada N+1 |
| 1 | Query tumbuh bersama baris — N+1 terdeteksi |
| 2 | Tidak bisa connect, seed gagal, atau ukuran kurang dari dua |

Exit code adalah vonisnya, jadi hasilnya lulus/gagal di layar.

## Catatan untuk demo

- Log statement GORM dimatikan di dalam alat ini. Tanpa itu, ratusan baris SQL
  mengubur vonisnya — persis hal yang harus tetap terbaca di proyektor.
- Jangan bungkus dengan `make`: exit code-nya adalah vonis, jadi make akan
  menambahkan `make: *** Error 1` tepat setelah vonis muncul.
