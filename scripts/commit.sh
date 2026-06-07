#!/usr/bin/env bash
set -euo pipefail

TYPES=(feat fix chore docs refactor test style)

echo ""
echo "=== Brewly Commit Helper ==="
echo ""

echo "Commit type:"
select TYPE in "${TYPES[@]}"; do
  [[ -n "$TYPE" ]] && break
done

read -rp "Scope (optional, e.g. auth, menu, orders) [skip]: " SCOPE

read -rp "Summary (imperative, present tense, < 72 chars): " SUMMARY

if [[ -z "$SUMMARY" ]]; then
  echo "Summary is required." >&2
  exit 1
fi

if [[ -n "$SCOPE" ]]; then
  MSG="${TYPE}(${SCOPE}): ${SUMMARY}"
else
  MSG="${TYPE}: ${SUMMARY}"
fi

echo ""
echo "Commit message: $MSG"
read -rp "Proceed? [y/N]: " CONFIRM
if [[ "${CONFIRM,,}" != "y" ]]; then
  echo "Aborted." >&2
  exit 1
fi

git commit -m "$MSG"
