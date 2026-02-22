#!/bin/bash
set -e

echo "Linting markdown files..."
npx -y markdownlint-cli "agents/**/*.md" "skills/**/*.md" "commands/**/*.md" "rules/**/*.md"
echo "Linting complete!"
