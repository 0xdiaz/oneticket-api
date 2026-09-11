# Scan keamanan

Menjalankan [OWASP ZAP](https://www.zaproxy.org/) terhadap API yang sedang hidup,
memakai `api/openapi.yaml` untuk tahu endpoint apa saja yang ada. Ia tidak
menebak dengan merayapi.

```bash
docker pull ghcr.io/zaproxy/zaproxy:stable   # ~1,1 GB, sekali saja
go run main.go &
./scripts/security/run.sh                    # passive, ~1,5 menit
./scripts/security/run.sh --active           # kirim payload serangan sungguhan
```

Laporan HTML dan JSON ditulis ke `scripts/security/reports/` (tidak di-commit).

## Kenapa ini perlu, padahal test Go sudah hijau

Yang dilihat ZAP hidup di luar handler: header yang ditambahkan framework,
informasi yang dibocorkan format error, setelan transport. Itu semua urusan
konfigurasi, dan konfigurasi justru yang tidak pernah dilihat unit test, karena
tidak ada fungsi yang salah.

## Yang ditemukan di kondisi awal repo

117 cek PASS, nol FAIL, dua peringatan:

```
WARN-NEW: A Server Error response code was returned by the server
          /api/v1/events/546058336831160984 (500)
WARN-NEW: X-Content-Type-Options Header Missing  x 6
```

Yang pertama adalah bug yang sama dengan temuan `scripts/apitest`, lewat jalan
yang sama sekali berbeda. Dua alat yang tidak saling kenal berhenti di titik yang
sama, dan itu yang bikin temuannya layak dipercaya.

Yang kedua tidak akan pernah ditemukan test Go mana pun di repo ini.

## Catatan jaringan

Scanner jalan di dalam container, jadi `localhost` di sana adalah container itu
sendiri. `run.sh` menulis salinan spec dengan host `host.docker.internal` ke
direktori kerja, spec aslinya tidak disentuh.

`--active` mengirim payload injection dan traversal sungguhan dan menulis ke
database. Jangan diarahkan ke apa pun yang datanya lo sayangi.
