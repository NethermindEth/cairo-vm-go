#!/bin/sh

cp githooks/pre-commit .git/hooks/pre-commit

# Make the pre-commit hook executable
chmod +x .git/hooks/pre-commit

echo "Pre-commit hook installed."
