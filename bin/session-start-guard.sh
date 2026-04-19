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
