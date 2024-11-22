#!/usr/bin/env ash

set -euox pipefail

stash_pop() {
  git stash pop $(git stash list --pretty="%gd %s" | grep "pre-lint-stash" | head -1 | awk "{print \$1}")
}

trap stash_pop INT TERM

git stash push -k -m "pre-lint-stash"

npx commitlint --last --verbose

npx commitlint --from origin/main --to HEAD --verbose

go mod tidy

golangci-lint run --out-format sarif --timeout 10m

set +e
markdownlint-cli2 **/*.md
MARKDOWNLINT_EXIT_CODE=$?
set -e

cat markdownlint-cli2-sarif.sarif
rm -f markdownlint-cli2-sarif.sarif

if [ "$MARKDOWNLINT_EXIT_CODE" -gt 0 ]; then
  exit "$MARKDOWNLINT_EXIT_CODE"
fi

yamllint -f parsable .

if [ -n "${GITHUB_ACTIONS:-}" ] && [ -n "$(git status --porcelain)" ]; then
  exit 1
fi

stash_pop
