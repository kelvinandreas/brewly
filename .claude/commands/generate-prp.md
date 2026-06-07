# /generate-prp

You are expanding an `INITIAL.md` (rough feature description) into a complete PRP at `PRPs/<kebab-name>.md`.

## Process

1. Read the INITIAL.md the user references.
2. Read `PRPs/templates/prp_base.md` for the section structure.
3. Read `PRPs/examples/prp_auth.md` to calibrate depth and tone — match it.
4. Read these for current-state context (always):
   - `memory-bank/architecture.md`
   - `memory-bank/database-schema.md`
   - `memory-bank/api-contracts.md`
5. Ask 2–3 sharp clarifying questions if any required input is missing. Don't ask filler questions.
6. Write the new PRP. Every section from the template must be filled. Do not implement.

## Quality bar

- File paths to create/modify are exact, not "create something in `usecase/`".
- Task breakdown is ordered and dependency-aware.
- Validation plan lists the exact commands (`make test`, manual smoke steps).
- Pseudocode covers the trickiest function only — not every function.
