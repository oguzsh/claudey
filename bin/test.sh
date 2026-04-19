#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

echo "Running cargo fmt --check..."
cargo fmt --all -- --check

echo "Running cargo clippy..."
cargo clippy --all-targets -- -D warnings

echo "Running cargo test..."
cargo test

echo "All checks passed!"
