#!/bin/bash
set -e

echo "Running validators..."
go run ./scripts/claudey-hooks validate-agents
go run ./scripts/claudey-hooks validate-commands
go run ./scripts/claudey-hooks validate-rules
go run ./scripts/claudey-hooks validate-skills
go run ./scripts/claudey-hooks validate-hooks

echo "All validations passed!"
