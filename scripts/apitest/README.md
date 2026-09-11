# Fuzzing kontrak API

Menjalankan [Schemathesis](https://schemathesis.readthedocs.io/) terhadap
`api/openapi.yaml`.

```bash
go run main.go &                    # butuh server hidup
./scripts/apitest/run.sh            # endpoint baca saja
./scripts/apitest/run.sh --all      # termasuk yang menulis, pakai DB sekali pakai
MAX_EXAMPLES=200 ./scripts/apitest/run.sh
SEED= ./scripts/apitest/run.sh      # acak, tiap run beda
```

## Soal seed

`run.sh` memakai seed yang dipin. Fuzzing itu acak, jadi tanpa itu jumlah
temuannya goyang antar-run: bug 500 muncul tiap kali, tapi temuan id negatif cuma
kadang-kadang. Kalau outputnya mau dibaca orang dari proyektor, hasil yang beda
tiap dijalankan itu mahal.

Jalankan dengan `SEED=` kosong kalau memang mau berburu yang belum pernah ketemu.
Seed yang dipakai selalu dicetak di ringkasan, jadi temuan baru bisa diulang.

Tidak ada kode test yang perlu ditulis. Setiap kasus dibangkitkan dari spec yang
sudah ada di repo: Schemathesis membaca tipe dan batas di spec, mengirim request
yang menurut spec sah, lalu memeriksa jawabannya terhadap apa yang spec janjikan.

Konsekuensinya yang menarik: **suite ini tumbuh sendiri.** Tambah satu field di
spec, jumlah kasus yang diuji naik tanpa ada yang mengetik test. Dan spec yang
melenceng dari kode langsung jadi kegagalan, jadi memperbarui spec berhenti jadi
pekerjaan dokumentasi.

## Yang ditemukan di kondisi awal repo

Empat temuan di dua endpoint, dalam 0,14 detik, sama tiap kali dijalankan:

```
GET /api/v1/events/9223372036854775808  ->  500
    unable to encode ... into binary format for int4 (OID 23)
```

Id di-parse sebagai `uint64` lalu dikirim ke kolom `int4`. Tidak ada orang yang
akan menulis test dengan angka itu, dan itulah gunanya alat ini: batas yang tidak
pernah terpikirkan.

Tiga sisanya: `data` di wrapper dideklarasikan `object` tapi `/events`
mengembalikan array; `/events/-165831` ditolak 400 padahal spec tidak melarang
angka negatif; `TRACE` dijawab 404 alih-alih 405.

Setiap temuan datang dengan perintah `curl`-nya, jadi tidak perlu percaya
laporannya, tinggal diulang sendiri.

## Instalasi

Tidak ada. `run.sh` memanggil Schemathesis lewat `uvx`, yang menariknya ke
environment sekali pakai. Repo ini tetap tidak punya dependency Python. Yang
dibutuhkan cuma `uv` (`brew install uv`).
