#!/bin/bash
set -e

echo "Building claudey binary..."
go build -o claudey ../scripts/claudey

