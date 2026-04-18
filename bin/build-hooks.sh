#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
echo "Building claudey binary..."
cargo build --release
install -m 0755 target/release/claudey bin/claudey
echo "Installed bin/claudey ($(ls -lh bin/claudey | awk '{print $5}'))"
