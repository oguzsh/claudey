#!/usr/bin/env bash
set -euo pipefail
HERE=$(cd "$(dirname "$0")" && pwd)
# shellcheck source=./lib.sh
source "$HERE/lib.sh"

REPO=$(cd "$HERE/../.." && pwd)
GUARD="$REPO/bin/session-start-guard.sh"

# -- Test 1: missing binary -> nudge + exit 1 ---------------------------------
T1=$(mk_fake_repo)
cp "$GUARD" "$T1/bin/session-start-guard.sh"
# Ensure bin/claudey does NOT exist in the fake repo.
run_capture "CLAUDE_PLUGIN_ROOT='$T1' bash '$T1/bin/session-start-guard.sh'"
assert_eq "$STATUS" "1"
assert_contains "$ERR" "hook binary not built yet"
assert_contains "$ERR" "bin/setup.sh"
rm -rf "$T1"
echo "  ok: missing binary -> nudge"

# -- Test 2: binary up-to-date -> exit 0 fast ---------------------------------
T2=$(mk_fake_repo)
cp "$GUARD" "$T2/bin/session-start-guard.sh"
touch "$T2/bin/claudey"
chmod +x "$T2/bin/claudey"
# Make bin/claudey newer than src/ by touching it last.
sleep 1
touch "$T2/bin/claudey"
run_capture "CLAUDE_PLUGIN_ROOT='$T2' bash '$T2/bin/session-start-guard.sh'"
assert_eq "$STATUS" "0"
assert_eq "$ERR" ""
rm -rf "$T2"
echo "  ok: up-to-date -> no-op"

# -- Test 3: src/ newer than binary -> rebuild via build-hooks.sh -------------
T3=$(mk_fake_repo)
cp "$GUARD" "$T3/bin/session-start-guard.sh"
touch "$T3/bin/claudey"
chmod +x "$T3/bin/claudey"
sleep 1
# Touch src/main.rs AFTER the binary to trigger the rebuild path.
touch "$T3/src/main.rs"
# Stub cargo so the PATH check passes (the fake build-hooks.sh doesn't call cargo,
# but the guard checks `command -v cargo` before invoking it).
STUB_BIN="$T3/stubbin"
mk_stub "$STUB_BIN/cargo"
run_capture "CLAUDE_PLUGIN_ROOT='$T3' PATH='$STUB_BIN:/usr/bin:/bin' bash '$T3/bin/session-start-guard.sh'"
assert_eq "$STATUS" "0"
assert_contains "$ERR" "source changed, rebuilding"
assert_contains "$ERR" "fake build-hooks: built bin/claudey"
rm -rf "$T3"
echo "  ok: stale source -> rebuild"

# -- Test 4: src/ newer AND cargo missing -> nudge ----------------------------
T4=$(mk_fake_repo)
cp "$GUARD" "$T4/bin/session-start-guard.sh"
touch "$T4/bin/claudey"
chmod +x "$T4/bin/claudey"
sleep 1
touch "$T4/src/main.rs"
# Empty PATH directory -> no cargo visible. Use a sentinel PATH with only coreutils.
run_capture "CLAUDE_PLUGIN_ROOT='$T4' PATH='/usr/bin:/bin' HOME='$T4' bash '$T4/bin/session-start-guard.sh'"
assert_eq "$STATUS" "1"
assert_contains "$ERR" "source changed but cargo is missing"
rm -rf "$T4"
echo "  ok: stale + no cargo -> nudge"
