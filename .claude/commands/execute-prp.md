# /execute-prp

You are implementing the PRP at the path the user gives you.

## Process

1. Read the PRP end-to-end. Re-read the Context section.
2. Re-read these for current state:
   - `memory-bank/architecture.md`
   - `memory-bank/database-schema.md`
   - `memory-bank/api-contracts.md`
3. Work the task breakdown in order. After each task:
   - run `make lint` for files in scope
   - run the relevant `make test-*` target
4. After the PRP is fully implemented:
   - update `memory-bank/database-schema.md` if any migration was added
   - update `memory-bank/api-contracts.md` if any endpoint changed
   - tick the matching items in `memory-bank/progress.md`
5. Stage the work and ask the user to run `make commit`.

## Hard rules during execution

- Stay inside the file structure declared in the PRP. If you need a new file outside scope, stop and ask.
- Never edit a merged migration.
- Never put business logic in handlers or fetch calls in components.
- Don't bypass the response envelope.
