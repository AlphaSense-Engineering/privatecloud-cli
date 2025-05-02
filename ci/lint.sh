#!/usr/bin/env ash

set -euox pipefail

stash_pop() {
  set +e
  git stash pop $(git stash list --pretty="%gd %s" | grep "pre-lint-stash" | head -1 | awk "{print \$1}")
  set -e
}

trap stash_pop INT TERM

git stash push -k -u -m "pre-lint-stash"

run() {
  set +e
  $1
  local EXIT_CODE=$?
  set -e

  if [ "$EXIT_CODE" -gt 0 ]; then
    stash_pop
    exit "$EXIT_CODE"
  fi
}

run "npx commitlint --last --verbose"
run "npx commitlint --from origin/main --to HEAD --verbose"

run "go mod tidy"
run "git diff --exit-code go.mod go.sum"

run "golangci-lint run --out-format sarif --timeout 10m"

set +e
markdownlint-cli2 **/*.md
MARKDOWNLINT_EXIT_CODE=$?
set -e

cat markdownlint-cli2-sarif.sarif
rm -f markdownlint-cli2-sarif.sarif

if [ "$MARKDOWNLINT_EXIT_CODE" -gt 0 ]; then
  stash_pop
  exit "$MARKDOWNLINT_EXIT_CODE"
fi

run "yamllint -f parsable ."

if [ -n "${GITHUB_ACTIONS:-}" ] && [ -n "$(git status --porcelain)" ]; then
  exit 1
fi

stash_pop
