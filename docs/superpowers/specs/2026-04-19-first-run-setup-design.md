# First-Run Setup Design

**Date:** 2026-04-19
**Status:** Approved (brainstorming)
**Next step:** Implementation plan via `superpowers:writing-plans`

## Goal

Make the `claudey` plugin ready-to-use the moment a user installs it, even though the plugin ships as Rust source that needs compilation. The first Claude Code session on a new machine should guide the user through a one-time setup; subsequent sessions should auto-rebuild silently when source changes, otherwise get out of the way.

## Constraints

- **Platforms:** macOS and Linux/WSL only. No Windows native.
- **Distribution:** Rust source only — no prebuilt binaries shipped. The binary at `bin/claudey` is gitignored and built per-machine.
- **Hooks are non-interactive:** Claude Code hooks receive JSON on stdin and write to stderr; there is no reliable TTY for `read -p` prompts inside a hook.
- **Chicken-and-egg:** The SessionStart hook itself invokes `bin/claudey session-start`. If the binary is missing, the hook cannot compile it from inside its own subcommand — the setup step must live in plain shell and run before the Rust binary is invoked.
- **No auto-install of Rust from inside a hook.** Rustup install is interactive, long-running, and network-heavy. It belongs in a user-run terminal session, not a hook.

## Architecture

Two new shell scripts, one line changed in `hooks/hooks.json`. Binary existence itself is the "setup done" marker — no separate `.setup-done` file.

### Roles

| Script | How it runs | Interactivity | Responsibilities |
|---|---|---|---|
| `bin/setup.sh` | Manually, by user in their terminal | Interactive — can prompt | If `cargo` missing → Y/n prompt, run `rustup` installer. Always finishes by invoking `bin/build-hooks.sh`. |
| `bin/session-start-guard.sh` | From the SessionStart hook, before `bin/claudey session-start` | Non-interactive — stderr only | If `bin/claudey` missing → stderr nudge, exit 1. Else if `src/` / `Cargo.toml` / `Cargo.lock` newer than binary → silent `cargo build --release` + install. Else exit 0 fast. |

`bin/build-hooks.sh` stays as the single place that runs `cargo build --release` and installs to `bin/claudey`. Both new scripts delegate to it — no duplicated build logic.

### hooks/hooks.json change

```diff
-"command": "\"${CLAUDE_PLUGIN_ROOT}/bin/claudey\" session-start"
+"command": "\"${CLAUDE_PLUGIN_ROOT}/bin/session-start-guard.sh\" && \"${CLAUDE_PLUGIN_ROOT}/bin/claudey\" session-start"
```

Only the SessionStart hook gets wrapped. Other hooks (PreToolUse / PostToolUse / Stop / etc.) trust the binary exists — in practice SessionStart always fires first on session open, so by the time any tool-triggered hook fires, the guard has already run.

### Guard state machine

```
bin/claudey missing                               → nudge to stderr, exit 1
bin/claudey exists, src/ newer than binary        → cargo build --release, install, exit 0
bin/claudey exists, src/ not newer                → exit 0 (stat-only, ~10ms)
cargo missing during silent rebuild (edge case)   → nudge to stderr, exit 1
```

On exit-1 the SessionStart hook fails visibly; Claude Code itself keeps running without the plugin's hooks.

## Script contents

### `bin/setup.sh`

```bash
#!/usr/bin/env bash
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

### `bin/session-start-guard.sh`

```bash
#!/usr/bin/env bash
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

if [[ -f "$HOME/.cargo/env" ]]; then
  # shellcheck disable=SC1091
  source "$HOME/.cargo/env"
fi

if [[ -n "$(find src Cargo.toml Cargo.lock -newer bin/claudey -print -quit 2>/dev/null)" ]]; then
  if ! command -v cargo >/dev/null 2>&1; then
    nudge "source changed but cargo is missing"
  fi
  echo "claudey: source changed, rebuilding..." >&2
  bash bin/build-hooks.sh >&2
fi
```

**Design choices baked in:**

- Rustup is invoked with `-y --default-toolchain stable` so it doesn't reprompt after the user already said Y.
- The rebuild trigger watches `src/`, `Cargo.toml`, `Cargo.lock`. It does *not* watch `tests/` (test changes shouldn't trigger a release rebuild) or `bin/*.sh` (shell changes don't affect the binary).
- `find ... -newer ... -print -quit` exits on first match for a fast path.
- `source "$HOME/.cargo/env"` is guarded by `[[ -f ... ]]` so it's a no-op when cargo came from apt/brew rather than rustup.

## Files to modify

- `hooks/hooks.json` — one-line SessionStart command change (above).
- `README.md` — "Quick Start" and "Manual Install" sections updated so first-time users see `bash bin/setup.sh` as the one-step install, replacing `bash bin/build-hooks.sh`. The existing `cp -r skills/ ...` commands for non-`/plugin install` users stay.
- `docs/binary.md` — add a short "First-run setup" section explaining the guard + setup script interplay. The existing "build per-machine" note still applies.
- `.gitignore` — no change required. Current exclusions are `bin/claudey` and `bin/claudey.exe` specifically; shell scripts in `bin/` are already tracked. Verify during implementation.

## Files not touched

- `bin/build-hooks.sh`, `bin/test.sh` — unchanged.
- All Rust source (`src/**`, `tests/**`) — unchanged. This is a shell/wiring change.
- All other entries in `hooks/hooks.json` — unchanged.

## Failure modes

| Scenario | Behavior | Recovery |
|---|---|---|
| Fresh clone, user hasn't run setup.sh | `bin/claudey` missing → stderr nudge, exit 1. SessionStart hook fails; Claude continues. | User runs `bash bin/setup.sh` in terminal. |
| User declined rustup prompt | `setup.sh` exits 1 early. `bin/claudey` never built. Next session → same nudge. | User re-runs `setup.sh` and accepts, or installs Rust by other means. |
| Rustup install fails (network, disk) | `setup.sh` exits non-zero via `set -e`. Partial `~/.cargo` may exist but unused. | User fixes root cause, re-runs `setup.sh`. Rustup installer is idempotent. |
| `cargo build` fails during silent rebuild | Guard exits non-zero via `set -e`. SessionStart chain breaks. User sees cargo error on stderr. | User inspects error, fixes source or `git checkout` a good commit. Manual: `bash bin/build-hooks.sh`. |
| Source pulled, old binary stale | Guard detects → silent `cargo build --release` on first SessionStart of that new source. | Automatic; 5-30 sec delay on that one session. |
| Source touched with no real change | Guard rebuilds unnecessarily. | Harmless — cargo's incremental cache keeps the rebuild at ~1-2 sec (no codegen). |
| `bin/claudey` deleted but cargo present | Missing-binary nudge path. `setup.sh` skips rustup (cargo present) and jumps to `build-hooks.sh`. | One manual command. |
| `git pull` mid-session | Guard fires on next SessionStart, not the current one. Current session keeps using old binary. | Acceptable — no hot-reload for running hooks anyway. |
| `bin/` not writable | `install -m 0755` fails, cargo error on stderr. | User fixes permissions. |

## Platform notes

- **macOS Gatekeeper / quarantine:** locally compiled binaries aren't quarantined. Irrelevant unless we later ship prebuilts.
- **WSL `$HOME/.cargo/env`:** sourced both in `setup.sh` (post-rustup) and in the guard. The guard's `source` is guarded by `[[ -f ... ]]`.
- **PATH in Claude Code hooks:** hooks may not inherit `~/.cargo/bin`. The explicit `source "$HOME/.cargo/env"` in the guard handles this.

## Explicitly out of scope

- Windows native (PowerShell / cmd).
- Auto-upgrading the Rust toolchain — `rustup update` is never invoked. A future `rust-toolchain.toml` pin would auto-resolve via rustup on build; not doing that now.
- Prebuilt binary fallback. If this plugin later ships releases, the distribution change is additive — this design doesn't preclude it.
- Offline / airgapped installs. `setup.sh` needs network for rustup; the guard needs network on the first `cargo build` to fetch crates.
- CI: CI should call `bin/build-hooks.sh` directly, not `setup.sh`. No hook guard involvement needed.

## Verification plan

1. Fresh clone on macOS → open Claude Code → SessionStart fires → nudge appears → run `bash bin/setup.sh` → re-open session → hook works.
2. Same on WSL/Ubuntu with no prior Rust install.
3. `touch src/main.rs` → next SessionStart silently rebuilds.
4. Intentionally break `src/main.rs` → next SessionStart shows cargo error on stderr and SessionStart hook fails, but Claude itself keeps running.
5. `rm bin/claudey` with cargo still present → nudge → `bash bin/setup.sh` re-runs and rebuilds without re-downloading rustup.
