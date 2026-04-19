#!/usr/bin/env bash
# Run every tests/shell/test_*.sh. Exit non-zero if any fails.
set -euo pipefail
cd "$(dirname "$0")"

fail=0
for t in test_*.sh; do
  [[ -f "$t" ]] || continue
  echo "== $t =="
  if bash "$t"; then
    echo "  PASS"
  else
    echo "  FAIL"
    fail=1
  fi
done
exit "$fail"
