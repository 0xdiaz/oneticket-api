#!/usr/bin/env bash
#
# Scan the API for common web vulnerabilities with OWASP ZAP.
#
# This runs zap-api-scan.py, which reads api/openapi.yaml to learn every
# endpoint rather than guessing by crawling. It is a passive scan by default:
# ZAP sends ordinary requests and inspects the responses for missing security
# headers, leaked server details, cookie flags, and similar. It does not try to
# break anything.
#
# What it is good at is the layer no Go test covers, because none of it lives
# in a handler: headers the framework adds, information the error format gives
# away, transport settings. Those are configuration questions, and
# configuration is exactly what unit tests never see.
#
# Requires a running server and the ZAP image (~1.5GB, pull it before a demo):
#   docker pull ghcr.io/zaproxy/zaproxy:stable
set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT_DIR="${OUT_DIR:-$REPO_ROOT/scripts/security/reports}"
IMAGE="ghcr.io/zaproxy/zaproxy:stable"

# The scanner runs in a container; localhost there is the container itself.
# host.docker.internal is how it reaches the server running on this machine.
TARGET_HOST="${TARGET_HOST:-host.docker.internal}"
TARGET_PORT="${TARGET_PORT:-8000}"
LOCAL_URL="http://localhost:${TARGET_PORT}"

# -a adds the active scan, which sends actual attack payloads (injection,
# traversal). Slower, and it writes to the database, so it belongs on a
# throwaway instance rather than the one on a projector.
MODE_FLAG=""
if [[ "${1:-}" == "--active" ]]; then
  MODE_FLAG="-a"
  echo "MODE: active scan. Mengirim payload serangan sungguhan ke $LOCAL_URL."
  echo "Jangan diarahkan ke apa pun yang datanya lo sayangi."
  shift
fi

if ! curl -sf --max-time 3 "$LOCAL_URL/health" >/dev/null 2>&1; then
  echo "Server tidak menjawab di $LOCAL_URL. Jalankan 'go run main.go' dulu." >&2
  exit 2
fi

if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
  echo "Image ZAP belum ada. Tarik dulu, ukurannya ~1.5GB:" >&2
  echo "  docker pull $IMAGE" >&2
  exit 2
fi

mkdir -p "$OUT_DIR"
STAMP="$(date +%Y%m%d-%H%M%S)"

# openapi.yaml declares http://localhost:8000, which is wrong from inside the
# container. A copy with the reachable host is written into the work directory
# instead of editing the real spec.
WORK="$OUT_DIR/.work"
mkdir -p "$WORK"
sed "s|http://localhost:8000|http://${TARGET_HOST}:${TARGET_PORT}|" \
  "$REPO_ROOT/api/openapi.yaml" > "$WORK/openapi.yaml"

echo "Target  : $LOCAL_URL (dilihat container sebagai http://${TARGET_HOST}:${TARGET_PORT})"
echo "Laporan : $OUT_DIR/zap-$STAMP.html"
echo

docker run --rm \
  -v "$WORK:/zap/wrk/:rw" \
  -t "$IMAGE" zap-api-scan.py \
  -t /zap/wrk/openapi.yaml \
  -f openapi \
  $MODE_FLAG \
  -r "report.html" \
  -J "report.json"
STATUS=$?

[[ -f "$WORK/report.html" ]] && mv "$WORK/report.html" "$OUT_DIR/zap-$STAMP.html"
[[ -f "$WORK/report.json" ]] && mv "$WORK/report.json" "$OUT_DIR/zap-$STAMP.json"

echo
case $STATUS in
  0) echo "AMAN: tidak ada peringatan di ambang yang dipakai." ;;
  1) echo "TEMUAN: ada peringatan level FAIL. Buka laporan HTML di atas." ;;
  2) echo "PERINGATAN: ada temuan level WARN, tidak ada FAIL." ;;
  *) echo "Scanner keluar dengan status $STATUS, kemungkinan gagal menjalankan scan." ;;
esac
exit $STATUS
