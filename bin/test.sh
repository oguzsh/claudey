#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
echo "Running cargo test..."
cargo test
echo "All tests passed!"
