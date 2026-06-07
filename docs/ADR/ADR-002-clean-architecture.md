# ADR-002 — Clean architecture with hard layer boundaries

## Context
Brewly is a small project today but will accumulate features (printer support, inventory, loyalty) post-v1. Without structural rules, business logic predictably ends up in handlers and persistence concerns leak everywhere — turning every new feature into a refactor.

## Decision
Adopt clean architecture with four layers and one direction of dependency:

`handler → usecase → domain ← repository`

Rules:
- `domain/` has zero external imports (std + uuid only). It declares entities, sentinel errors, and `XxxRepository` interfaces.
- `usecase/` orchestrates. Depends only on `domain.*Repository`. Returns domain types or sentinel errors.
- `repository/` implements `domain.*Repository`. GORM lives here only.
- `handler/` is HTTP only. Decodes → calls usecase → encodes. No business logic.

Errors wrap context on every layer crossing: `fmt.Errorf("usecase.CreateOrder: %w", err)`.

## Consequences
- Pros: usecases are trivially testable with mocked repos; swapping ORM or HTTP router is mechanical; new contributors learn one repository pattern and stay productive.
- Cons: more files per feature than a flat `service.go` style; can feel ceremonial for trivial CRUD.
- Mitigation: provide canonical examples in `examples/go-handler.go` and `examples/go-usecase.go`, plus a `go-clean-arch` project skill that activates when editing these directories.
