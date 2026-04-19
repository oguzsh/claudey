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
