#!/usr/bin/env bash
set -euo pipefail
HERE=$(cd "$(dirname "$0")" && pwd)
# shellcheck source=./lib.sh
source "$HERE/lib.sh"

REPO=$(cd "$HERE/../.." && pwd)
SETUP="$REPO/bin/setup.sh"

# -- Test 1: cargo already present -> skip rustup, call build-hooks.sh --------
T1=$(mk_fake_repo)
cp "$SETUP" "$T1/bin/setup.sh"
STUB_BIN="$T1/stubbin"
mk_stub "$STUB_BIN/cargo"
# curl stub with a sentinel name we can grep for — if rustup branch is wrongly
# taken, we'll see "curl-called" in output.
cat > "$STUB_BIN/curl" <<'EOF'
#!/usr/bin/env bash
echo "curl-called-WRONG $*"
exit 0
EOF
chmod +x "$STUB_BIN/curl"
run_capture "PATH='$STUB_BIN:/usr/bin:/bin' bash '$T1/bin/setup.sh'"
assert_eq "$STATUS" "0"
assert_contains "$OUT" "fake build-hooks: built bin/claudey"
assert_contains "$OUT" "setup complete"
if [[ "$OUT" == *"curl-called-WRONG"* || "$ERR" == *"curl-called-WRONG"* ]]; then
  echo "FAIL: curl was invoked even though cargo was present" >&2
  exit 1
fi
rm -rf "$T1"
echo "  ok: cargo present -> build only"

# -- Test 2: cargo missing, user declines -> exit 1 ---------------------------
T2=$(mk_fake_repo)
cp "$SETUP" "$T2/bin/setup.sh"
# No cargo on PATH, HOME points at tempdir so ~/.cargo/env doesn't exist.
run_capture "PATH='/usr/bin:/bin' HOME='$T2' bash '$T2/bin/setup.sh' <<< 'n'"
assert_eq "$STATUS" "1"
assert_contains "$OUT" "Rust toolchain not found"
assert_contains "$OUT" "Aborted"
rm -rf "$T2"
echo "  ok: cargo missing + decline -> exit 1"

# -- Test 3: cargo missing, user accepts -> curl invoked, then build ----------
T3=$(mk_fake_repo)
cp "$SETUP" "$T3/bin/setup.sh"
STUB_BIN="$T3/stubbin"
mkdir -p "$STUB_BIN"
# curl stub simulates rustup installer: outputs shell script that will be piped to sh.
cat > "$STUB_BIN/curl" <<EOF
#!/usr/bin/env bash
# Output shell code that the piped sh will execute.
cat <<'SHELL'
mkdir -p "$T3/.cargo/bin"
cat > "$T3/.cargo/bin/cargo" <<'CARGO'
#!/usr/bin/env bash
echo "[rustup-installed cargo] \$*"
exit 0
CARGO
chmod +x "$T3/.cargo/bin/cargo"
cat > "$T3/.cargo/env" <<'ENV'
export PATH="\$PATH:$T3/.cargo/bin"
ENV
echo "curl-rustup-installed"
SHELL
exit 0
EOF
chmod +x "$STUB_BIN/curl"
run_capture "PATH='$STUB_BIN:/usr/bin:/bin' HOME='$T3' bash '$T3/bin/setup.sh' <<< 'y'"
assert_eq "$STATUS" "0"
assert_contains "$OUT" "Rust toolchain not found"
assert_contains "$OUT" "curl-rustup-installed"
assert_contains "$OUT" "fake build-hooks: built bin/claudey"
assert_contains "$OUT" "setup complete"
rm -rf "$T3"
echo "  ok: cargo missing + accept -> rustup + build"
