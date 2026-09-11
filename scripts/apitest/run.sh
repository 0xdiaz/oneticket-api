#!/usr/bin/env bash
#
# Fuzz the API against its own OpenAPI spec.
#
# Schemathesis reads api/openapi.yaml, generates requests that the spec says
# are valid, and checks the answers against what the spec promises. It needs
# no test code: every case it runs comes from the contract that already exists
# in the repo.
#
# That is the point worth making. The spec was written to document the API.
# Pointing a fuzzer at it turns the same file into a test suite, and the suite
# grows by itself every time the spec gains a field.
#
# Requires a running server. Start one with `go run main.go` first.
set -uo pipefail

BASE_URL="${BASE_URL:-http://localhost:8000}"
SPEC="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)/api/openapi.yaml"
MAX_EXAMPLES="${MAX_EXAMPLES:-25}"

# Read-only by default. The write endpoints register users, revoke sessions and
# mail password resets, so fuzzing them needs a throwaway database rather than
# whatever the machine happens to be pointing at. Pass --all to include them.
INCLUDE='^/api/v1/events'
if [[ "${1:-}" == "--all" ]]; then
  INCLUDE='.'
  echo "MODE: seluruh spec, termasuk endpoint yang menulis. Pastikan database ini sekali pakai."
  shift
fi

if ! curl -sf --max-time 3 "$BASE_URL/health" >/dev/null 2>&1; then
  echo "Server tidak menjawab di $BASE_URL. Jalankan 'go run main.go' dulu." >&2
  exit 2
fi

# uvx fetches Schemathesis into a throwaway environment, so nothing is
# installed permanently and the repo keeps no Python dependency.
if ! command -v uvx >/dev/null 2>&1; then
  echo "uvx tidak ada. Pasang uv: brew install uv" >&2
  exit 2
fi

echo "Spec   : $SPEC"
echo "Target : $BASE_URL"
echo "Scope  : $INCLUDE"
echo

uvx --from schemathesis st run "$SPEC" \
  --url "$BASE_URL" \
  --include-path-regex "$INCLUDE" \
  --max-examples "$MAX_EXAMPLES" \
  "$@"
STATUS=$?

echo
if [[ $STATUS -eq 0 ]]; then
  echo "AMAN: tidak ada jawaban yang melanggar spec."
else
  echo "TEMUAN: lihat blok FAILURES di atas. Tiap temuan punya perintah curl untuk"
  echo "mengulangnya, jadi tidak perlu percaya laporannya, tinggal jalankan sendiri."
fi
exit $STATUS
