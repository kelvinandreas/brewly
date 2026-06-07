# PRP — <Feature name>

## Goal
<One paragraph: what shipping this feature accomplishes and the user-visible outcome.>

## Context
<Load-bearing facts. Cite memory-bank/* files. Note any constraints from ADRs.>

## File structure
### To create
- `backend/internal/...`
- `frontend/src/...`

### To modify
- `backend/internal/...`
- `memory-bank/api-contracts.md` — add new endpoints
- `memory-bank/database-schema.md` — if migration added
- `memory-bank/progress.md`

## Task breakdown
1. <step>
2. <step>
3. …

## Pseudocode
<For the trickiest function only.>

## Validation plan
- `make lint`
- `make test-backend` / `make test-frontend`
- Manual smoke: <exact steps>

## Out of scope
<What's deliberately deferred.>
