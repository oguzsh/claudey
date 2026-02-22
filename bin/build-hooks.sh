#!/bin/bash
set -e

echo "Building claudey binary..."
go build -o bin/claudey ./scripts/claudey

