# Go backend rules

Globs: `**/*.go`

## Layering
- `handler/` → `usecase/` → `domain/` ← `repository/`. Never cross.
- `domain/` has no third-party imports except `github.com/google/uuid`.
- GORM lives in `repository/` only.

## Errors
- Sentinel errors in `internal/domain/errors.go` (e.g., `ErrOrderNotFound`).
- Wrap on layer crossings: `fmt.Errorf("usecase.CreateOrder: %w", err)`.
- Handler maps `errors.Is(err, domain.ErrXxx)` → HTTP status code.
- Never `panic` in normal flow. Recover middleware catches truly unexpected.

## Tests
- Table-driven with `t.Run(name, func(t *testing.T) { ... })`.
- Usecase tests use mocks for repositories (interfaces from `domain/`).
- Repository tests use a real Postgres testcontainer.

## Style
- Godoc on every exported func / type.
- Receiver name: first letter of struct lowercase, e.g., `o *OrderUsecase`.
- No magic numbers — put constants in `internal/domain/constants.go`.
- HTTP responses through `pkg/response.OK(w, data)` and `pkg/response.Error(w, status, code, msg)`.

## Validation
- Request DTOs in the handler file. Tag with `validate:"required,..."`.
- One shared `*validator.Validate` injected via `cmd/api/main.go`.
- Field errors mapped to `details: [{field, message}]` in the response envelope.
