#!/bin/bash
set -e

echo "Building claudey-hooks binary..."
go build -o hooks/bin/claudey-hooks ./scripts/claudey-hooks

echo "Running linter..."
./bin/lint.sh
