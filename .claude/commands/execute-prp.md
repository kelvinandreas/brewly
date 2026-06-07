## Process

1. Read the PRP end-to-end. Re-read the Context section.
2. Re-read these for current state:
   - `memory-bank/architecture.md`
   - `memory-bank/database-schema.md`
   - `memory-bank/api-contracts.md`
3. Work the task breakdown in order. **After each task:**
   - run `make lint` for files in scope
   - run the relevant `make test-*` target
   - **Stage the task's files and commit atomically** (one logical change = one commit)
   - Ask user: "Task X done. Ready to commit?" → they run `make commit` or you run it
4. After the PRP is fully implemented:
   - update `memory-bank/database-schema.md` if any migration was added
   - update `memory-bank/api-contracts.md` if any endpoint changed
   - tick the matching items in `memory-bank/progress.md`
   - **Final commit** for memory-bank updates: `chore(docs): update memory-bank after [feature name]`
5. All work pushed — PRP complete.

## Atomic commit rules

- One logical change per commit (one task = one commit).
- Never merge unrelated files into one commit.
- If a task touches backend + frontend + migration, that's still **one task = one commit** (they're coupled).
- Example: "Task 2: Create JWTUsecase" → one commit `feat(auth): add JWT signing and verification`.
