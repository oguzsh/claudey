#!/bin/bash
set -e

echo "Running validations..."
go run ./scripts/claudey validate-agents
go run ./scripts/claudey validate-commands
go run ./scripts/claudey validate-rules
go run ./scripts/claudey validate-skills
go run ./scripts/claudey validate-hooks
echo "All validations passed!"
