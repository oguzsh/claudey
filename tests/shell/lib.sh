#!/usr/bin/env bash
# Test helpers for shell integration tests.
set -euo pipefail

# Create an isolated fake repo that mimics the claudey layout.
# Prints the tempdir path on stdout.
mk_fake_repo() {
  local root
  root=$(mktemp -d)
  mkdir -p "$root/bin" "$root/src" "$root/target/release"
  printf 'fn main() {}\n' > "$root/src/main.rs"
  printf '[package]\nname = "claudey"\nversion = "0.1.0"\nedition = "2021"\n' > "$root/Cargo.toml"
  printf '# lock\n' > "$root/Cargo.lock"
  # Fake build-hooks.sh: touches bin/claudey so the "setup complete" path works
  # without actually compiling.
  cat > "$root/bin/build-hooks.sh" <<'FAKE'
#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
touch bin/claudey
chmod +x bin/claudey
echo "fake build-hooks: built bin/claudey"
FAKE
  chmod +x "$root/bin/build-hooks.sh"
  printf '%s' "$root"
}

# Create a stub binary at $1 that echoes its name + args and exits 0.
mk_stub() {
  local path=$1
  mkdir -p "$(dirname "$path")"
  cat > "$path" <<'STUB'
#!/usr/bin/env bash
echo "[stub $(basename "$0")] $*"
exit 0
STUB
  chmod +x "$path"
}

# Assert that $1 equals $2; print both on failure and exit 1.
assert_eq() {
  if [[ "$1" != "$2" ]]; then
    printf 'FAIL: expected=%q actual=%q\n' "$2" "$1" >&2
    exit 1
  fi
}

# Assert that $1 contains substring $2; print both on failure and exit 1.
assert_contains() {
  if [[ "$1" != *"$2"* ]]; then
    printf 'FAIL: %q\n  does not contain %q\n' "$1" "$2" >&2
    exit 1
  fi
}

# Run $1 (a command string) and capture: STATUS=exit code, OUT=stdout, ERR=stderr.
# Usage:
#   run_capture 'some-command arg1 arg2 <<< "input"'
run_capture() {
  local cmd=$1
  local out_file err_file
  out_file=$(mktemp)
  err_file=$(mktemp)
  set +e
  eval "$cmd" >"$out_file" 2>"$err_file"
  STATUS=$?
  set -e
  OUT=$(cat "$out_file")
  ERR=$(cat "$err_file")
  rm -f "$out_file" "$err_file"
}
