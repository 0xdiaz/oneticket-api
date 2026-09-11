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

# Fuzzing is random, so the number of findings moves between runs: the
# out-of-range id that returns 500 turns up every time, the negative id only
# sometimes. A pinned seed makes the run reproducible, which matters when the
# output is being read off a projector. Set SEED= (empty) to let it roam.
SEED="${SEED-233638011671589819716048247251880411702}"

# Read-only by default. The write endpoints register users, revoke sessions and
# mail password resets, so fuzzing them needs a throwaway database rather than
# whatever the machine happens to be pointing at. Pass --all to include them.
#
# The checkout path matches INCLUDE too, and it sells real tickets. It is
# excluded separately rather than by narrowing INCLUDE, because INCLUDE is also
# what carries the fuzzer to GET /api/v1/events/{id} -- the endpoint whose
# out-of-range-id 500 this tool is meant to keep finding. Narrowing the include
# pattern would silence that finding as a side effect.
INCLUDE='^/api/v1/events'
EXCLUDE='/purchase$'
if [[ "${1:-}" == "--all" ]]; then
  INCLUDE='.'
  EXCLUDE=''
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
if [[ -n "$EXCLUDE" ]]; then
  echo "Scope  : $INCLUDE (kecuali $EXCLUDE)"
else
  echo "Scope  : $INCLUDE"
fi
echo

SEED_ARG=()
[[ -n "$SEED" ]] && SEED_ARG=(--seed "$SEED")

EXCLUDE_ARG=()
[[ -n "$EXCLUDE" ]] && EXCLUDE_ARG=(--exclude-path-regex "$EXCLUDE")

uvx --from schemathesis st run "$SPEC" \
  --url "$BASE_URL" \
  --include-path-regex "$INCLUDE" \
  "${EXCLUDE_ARG[@]}" \
  --max-examples "$MAX_EXAMPLES" \
  "${SEED_ARG[@]}" \
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
