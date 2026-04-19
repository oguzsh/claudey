# First-Run Setup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the plugin ready-to-use after `/plugin install` by auto-building the Rust binary on first SessionStart (guided on first run, silent on subsequent source changes).

**Architecture:** Two new Bash scripts. `bin/setup.sh` runs manually in the user's terminal on first install — prompts before installing rustup if needed, then delegates to `bin/build-hooks.sh`. `bin/session-start-guard.sh` runs from the SessionStart hook — nudges the user to run `setup.sh` on missing binary, auto-rebuilds silently when `src/` / `Cargo.toml` / `Cargo.lock` are newer than `bin/claudey`. Binary existence itself is the "setup done" marker; no extra marker file.

**Tech Stack:** Bash (POSIX-ish, bash 3.2+ for macOS compatibility), `cargo` + `rustup` for the Rust toolchain. Tests are hand-rolled Bash assertion scripts under `tests/shell/` (no external test framework — keeps the dependency surface zero, matches the user's stdlib-first preference).

**Source spec:** `docs/superpowers/specs/2026-04-19-first-run-setup-design.md`

---

## Context

- `hooks/hooks.json` currently wires SessionStart to `"${CLAUDE_PLUGIN_ROOT}/bin/claudey" session-start`. Chicken-and-egg: if `bin/claudey` doesn't exist, this hook itself cannot build it. The guard script lives *before* the `&&` in the SessionStart command.
- `bin/build-hooks.sh` is the existing build script; it runs `cargo build --release` then `install -m 0755 target/release/claudey bin/claudey`. Both new scripts delegate to it — no duplicated build logic.
- `bin/claudey` is gitignored (tracked exclusions are exactly `bin/claudey` and `bin/claudey.exe`; shell scripts under `bin/` remain tracked).
- Claude Code hooks receive JSON on stdin and write to stderr — **no reliable TTY** for `read -p` prompts inside a hook. All interactive prompting lives in `setup.sh`, which the user runs in their terminal.
- Platforms in scope: macOS and Linux/WSL. No Windows native.

---

## Task 1: Scaffold shell test harness

**Files:**
- Create: `tests/shell/lib.sh`
- Create: `tests/shell/run.sh`

- [ ] **Step 1: Create `tests/shell/lib.sh`**

```bash
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
```

- [ ] **Step 2: Create `tests/shell/run.sh`**

```bash
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
```

- [ ] **Step 3: Make both executable**

```bash
chmod +x tests/shell/run.sh
# lib.sh is sourced, not executed, so its mode is advisory.
chmod +x tests/shell/lib.sh
```

- [ ] **Step 4: Verify harness runs with zero tests**

Run: `bash tests/shell/run.sh`
Expected: empty output (no `test_*.sh` files yet), exit code 0.

- [ ] **Step 5: Commit**

```bash
git add tests/shell/lib.sh tests/shell/run.sh
git commit -m "test: scaffold shell test harness for setup scripts"
```

---

## Task 2: Implement `bin/session-start-guard.sh`

**Why this one first:** It's the simpler of the two scripts and has no `rustup`/`read -p` concerns to mock. Getting it in first gives us a working guard to validate the hook wiring in Task 4 before touching `setup.sh`.

**Files:**
- Create: `tests/shell/test_guard.sh`
- Create: `bin/session-start-guard.sh`

- [ ] **Step 1: Write the failing tests**

Create `tests/shell/test_guard.sh`:

```bash
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
```

Make it executable:

```bash
chmod +x tests/shell/test_guard.sh
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `bash tests/shell/run.sh`
Expected: `== test_guard.sh ==` followed by failure — `bin/session-start-guard.sh` doesn't exist yet, so `cp "$GUARD" ...` fails.

- [ ] **Step 3: Implement `bin/session-start-guard.sh`**

```bash
#!/usr/bin/env bash
# Runs from the SessionStart hook, before `bin/claudey session-start`.
# Non-interactive: prints to stderr only, never prompts.
set -euo pipefail
cd "$(dirname "$0")/.."

nudge() {
  echo "claudey: $1" >&2
  echo "claudey: run \`bash \"\${CLAUDE_PLUGIN_ROOT:-$(pwd)}/bin/setup.sh\"\` in your terminal." >&2
  exit 1
}

if [[ ! -x bin/claudey ]]; then
  nudge "hook binary not built yet"
fi

# Hooks may not inherit ~/.cargo/bin. Source rustup's env if present.
if [[ -f "$HOME/.cargo/env" ]]; then
  # shellcheck disable=SC1091
  source "$HOME/.cargo/env"
fi

# Rebuild if any watched source is newer than the binary.
# `find -newer <target> -print -quit` exits on first match for speed.
if [[ -n "$(find src Cargo.toml Cargo.lock -newer bin/claudey -print -quit 2>/dev/null)" ]]; then
  if ! command -v cargo >/dev/null 2>&1; then
    nudge "source changed but cargo is missing"
  fi
  echo "claudey: source changed, rebuilding..." >&2
  bash bin/build-hooks.sh >&2
fi
```

Make it executable:

```bash
chmod +x bin/session-start-guard.sh
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `bash tests/shell/run.sh`
Expected:
```
== test_guard.sh ==
  ok: missing binary -> nudge
  ok: up-to-date -> no-op
  ok: stale source -> rebuild
  ok: stale + no cargo -> nudge
  PASS
```

- [ ] **Step 5: Commit**

```bash
git add bin/session-start-guard.sh tests/shell/test_guard.sh
git commit -m "feat(hooks): add session-start-guard.sh for first-run + rebuild"
```

---

## Task 3: Implement `bin/setup.sh`

**Files:**
- Create: `tests/shell/test_setup.sh`
- Create: `bin/setup.sh`

- [ ] **Step 1: Write the failing tests**

Create `tests/shell/test_setup.sh`:

```bash
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
# curl stub simulates rustup installer: creates ~/.cargo/env + a cargo shim.
cat > "$STUB_BIN/curl" <<EOF
#!/usr/bin/env bash
# Pretend to download and pipe rustup-init into sh.
# Our sh stub (next arg chain) creates a fake cargo env.
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
```

Make it executable:

```bash
chmod +x tests/shell/test_setup.sh
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `bash tests/shell/run.sh`
Expected: `== test_setup.sh ==` fails — `bin/setup.sh` doesn't exist yet, so `cp "$SETUP" ...` errors out.

- [ ] **Step 3: Implement `bin/setup.sh`**

```bash
#!/usr/bin/env bash
# One-time setup: ensures Rust toolchain is present, then builds bin/claudey.
# Run manually in your terminal after installing the claudey plugin.
set -euo pipefail
cd "$(dirname "$0")/.."

case "$(uname -s)" in
  Darwin|Linux) ;;
  *) echo "claudey: unsupported OS: $(uname -s)" >&2; exit 1 ;;
esac

if ! command -v cargo >/dev/null 2>&1; then
  echo "claudey: Rust toolchain not found."
  echo "rustup will download & install Rust (~300 MB, ~2-5 min)."
  read -r -p "Install now? [Y/n] " ans
  case "${ans:-Y}" in
    Y|y|yes|"") ;;
    *) echo "Aborted. Install Rust from https://rustup.rs then re-run this script."; exit 1 ;;
  esac
  curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain stable
  # shellcheck disable=SC1090
  source "$HOME/.cargo/env"
fi

bash bin/build-hooks.sh
echo "claudey: setup complete - bin/claudey is ready."
```

Make it executable:

```bash
chmod +x bin/setup.sh
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `bash tests/shell/run.sh`
Expected:
```
== test_guard.sh ==
  ok: missing binary -> nudge
  ok: up-to-date -> no-op
  ok: stale source -> rebuild
  ok: stale + no cargo -> nudge
  PASS
== test_setup.sh ==
  ok: cargo present -> build only
  ok: cargo missing + decline -> exit 1
  ok: cargo missing + accept -> rustup + build
  PASS
```

- [ ] **Step 5: Commit**

```bash
git add bin/setup.sh tests/shell/test_setup.sh
git commit -m "feat(bin): add setup.sh for first-run rustup + build"
```

---

## Task 4: Wire SessionStart hook to use the guard

**Files:**
- Modify: `hooks/hooks.json` (SessionStart matcher block)

- [ ] **Step 1: Read current SessionStart entry**

Run: `grep -A 10 '"SessionStart"' hooks/hooks.json`
Expected output (current state):
```
"SessionStart": [
  {
    "matcher": "*",
    "hooks": [
      {
        "type": "command",
        "command": "\"${CLAUDE_PLUGIN_ROOT}/bin/claudey\" session-start"
      }
    ]
  }
],
```

- [ ] **Step 2: Replace the SessionStart command line**

Edit `hooks/hooks.json`. Replace the single line:

```
"command": "\"${CLAUDE_PLUGIN_ROOT}/bin/claudey\" session-start"
```

with:

```
"command": "\"${CLAUDE_PLUGIN_ROOT}/bin/session-start-guard.sh\" && \"${CLAUDE_PLUGIN_ROOT}/bin/claudey\" session-start"
```

**Important:** this change applies to the `"SessionStart"` block only. Do not modify other hook entries — PreToolUse, PostToolUse, Stop, etc. still invoke `bin/claudey` directly.

- [ ] **Step 3: Validate JSON parses**

Run: `python3 -c 'import json, sys; json.load(open("hooks/hooks.json")); print("ok")'`
Expected: `ok`

- [ ] **Step 4: Spot-check the change**

Run: `grep -A 10 '"SessionStart"' hooks/hooks.json`
Expected: the SessionStart block's `command` now has `session-start-guard.sh &&` before the binary invocation. No other hook entries changed.

- [ ] **Step 5: Commit**

```bash
git add hooks/hooks.json
git commit -m "chore(hooks): wrap SessionStart with session-start-guard.sh"
```

---

## Task 5: Update README.md for the new one-step install

**Files:**
- Modify: `README.md` (Quick Start and Manual Install sections)

- [ ] **Step 1: Read current Manual Install section**

Run: `sed -n '29,46p' README.md`
Expected output:
```
## Manual Install

```bash
git clone https://github.com/oguzsh/claudey.git
cd claudey

# Build bin/claudey locally (Rust toolchain required)
bash bin/build-hooks.sh

# Copy what you need to ~/.claude/
cp -r skills/ ~/.claude/skills/
cp -r commands/ ~/.claude/commands/
cp -r rules/ ~/.claude/rules/
cp -r hooks/ ~/.claude/hooks/
```

`bin/claudey` is not checked into git -- each machine builds its own binary from source.
```

- [ ] **Step 2: Replace the Manual Install block**

Replace lines 29–46 (the section starting at `## Manual Install` through the "not checked into git" sentence) with:

```
## Manual Install

```bash
git clone https://github.com/oguzsh/claudey.git
cd claudey

# One-time setup: prompts to install Rust via rustup if missing, then builds bin/claudey
bash bin/setup.sh

# Copy what you need to ~/.claude/
cp -r skills/ ~/.claude/skills/
cp -r commands/ ~/.claude/commands/
cp -r rules/ ~/.claude/rules/
cp -r hooks/ ~/.claude/hooks/
```

`bin/claudey` is not checked into git — each machine builds its own binary from source via `bash bin/setup.sh`. After the first build, new sessions auto-rebuild whenever `src/` changes (no prompt, streamed to stderr).
```

- [ ] **Step 3: Verify no stale `build-hooks.sh` references remain in user-facing install flow**

Run: `grep -n 'build-hooks\.sh' README.md`
Expected: no matches (users see `bin/setup.sh` now; `build-hooks.sh` is an internal detail documented in `docs/binary.md`).

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs(readme): point install flow at bin/setup.sh"
```

---

## Task 6: Update docs/binary.md with the first-run section

**Files:**
- Modify: `docs/binary.md` (append a new section)

- [ ] **Step 1: Read the end of docs/binary.md**

Run: `tail -20 docs/binary.md`
Expected: ends with existing content about the binary's subcommand dispatcher or build flow.

- [ ] **Step 2: Append the First-Run Setup section**

Append to `docs/binary.md`:

```markdown

## First-Run Setup

The plugin ships as Rust source, not a prebuilt binary, so `bin/claudey` has to be built per-machine. Two shell scripts mediate this:

- **`bin/setup.sh`** — run manually in your terminal after installing the plugin. If `cargo` is missing, prompts to install Rust via `rustup` (one-line Y/n). Then delegates to `bin/build-hooks.sh` to build and install the binary. Run once per machine.
- **`bin/session-start-guard.sh`** — invoked automatically at the start of every Claude Code session, wrapped around `bin/claudey session-start` in `hooks/hooks.json`. If the binary is missing, prints a one-line nudge to run `bin/setup.sh` and exits (the SessionStart hook fails visibly; Claude itself keeps running). If the binary exists but `src/` / `Cargo.toml` / `Cargo.lock` are newer, silently rebuilds via `bin/build-hooks.sh` and streams cargo's output to stderr.

The binary's own existence is the "setup done" marker — no separate flag file. Manual rebuilds still work: `bash bin/build-hooks.sh` any time.

**Platform support:** macOS and Linux/WSL. Windows native is not supported.
```

- [ ] **Step 3: Verify section renders**

Run: `grep -A 2 '## First-Run Setup' docs/binary.md`
Expected: the heading plus the opening paragraph appear.

- [ ] **Step 4: Commit**

```bash
git add docs/binary.md
git commit -m "docs(binary): document first-run setup and guard"
```

---

## Task 7: End-to-end verification

**Files:** none modified — this task validates real behavior.

- [ ] **Step 1: Run the full shell test suite**

Run: `bash tests/shell/run.sh`
Expected: both test files pass (total 7 cases), exit 0.

- [ ] **Step 2: Verify `.gitignore` didn't accidentally exclude the new scripts**

Run: `git check-ignore -v bin/setup.sh bin/session-start-guard.sh || echo "not ignored"`
Expected: `not ignored` (both files should be tracked — only `bin/claudey` and `bin/claudey.exe` are excluded).

- [ ] **Step 3: Dry-run the guard against the current repo**

Run: `bash bin/session-start-guard.sh`
Expected (binary already built): exits 0 silently.
Expected (if binary was deleted): stderr prints "hook binary not built yet" + the setup nudge, exits 1.

- [ ] **Step 4: Simulate stale-source rebuild path**

```bash
# Make bin/claudey older than src/ so guard detects staleness.
touch -d '1 hour ago' bin/claudey
bash bin/session-start-guard.sh
```

Expected: stderr shows `claudey: source changed, rebuilding...` followed by cargo's build output, then `bin/claudey` is up-to-date again and the script exits 0.

- [ ] **Step 5: Validate hooks.json is still valid**

Run: `python3 -c 'import json; d=json.load(open("hooks/hooks.json")); print("ok:", len(d["hooks"]["SessionStart"][0]["hooks"]), "SessionStart hook(s)")'`
Expected: `ok: 1 SessionStart hook(s)` (or however many the file already had).

- [ ] **Step 6: Real Claude Code session smoke test** *(manual)*

Install the updated plugin locally (or rely on it already being installed via the marketplace). Start a new Claude Code session in this repo:

1. With `bin/claudey` present and fresh: SessionStart completes normally, no guard output visible to the user.
2. Delete `bin/claudey`, start a new session: the session-start hook logs the nudge message in hook diagnostics (check `~/.claude/logs/` or the IDE's hook output pane); Claude itself still starts normally.
3. Run `bash bin/setup.sh` in the terminal, then start another new session: hook works again.

If step 2 shows anything other than the nudge (e.g., cryptic "command not found"), revisit the guard's error message.

- [ ] **Step 7: Commit (no changes expected, but clean up if anything drifted)**

```bash
git status   # Expected: clean
```

If clean, skip. If something drifted, inspect and commit with an explanatory message.

---

## Verification summary

After all tasks complete:

1. `bash tests/shell/run.sh` passes (7 test cases across guard + setup).
2. `bin/session-start-guard.sh` is wired into `hooks/hooks.json` for SessionStart only.
3. `bin/setup.sh` works as the one-command install; README documents it.
4. `docs/binary.md` explains the guard/setup interplay.
5. Manual smoke test: fresh clone → `bash bin/setup.sh` → Claude Code session works.
6. Manual smoke test: delete `bin/claudey` → new Claude Code session shows the nudge.
7. Manual smoke test: `touch src/main.rs` → next session silently rebuilds.

---

## Reference

- Source spec: `docs/superpowers/specs/2026-04-19-first-run-setup-design.md`
- Existing build script (delegated to by both new scripts): `bin/build-hooks.sh`
- Hook wiring file touched by Task 4: `hooks/hooks.json` (SessionStart block)
- Plugin-root env var used in the nudge message: `CLAUDE_PLUGIN_ROOT` (set by Claude Code when running hooks)
