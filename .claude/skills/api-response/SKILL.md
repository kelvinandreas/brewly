---
name: api-response
description: Brewly's standard API response envelope, HTTP status mapping, validation error shape, and "never leak internals" rule. Triggers when writing HTTP handlers or any error-returning Go code.
---

## Envelope

Success:
```json
{ "success": true, "data": "<object|array>", "message": "<optional>" }
```

Failure:
```json
{ "success": false, "error": "<machine code>", "details": ["<optional>"] }
```

Always route through `pkg/response`:

```go
response.OK(w, order)
response.Error(w, http.StatusUnprocessableEntity, "empty_order", "order must have items")
response.ValidationErrors(w, []response.FieldError{{Field: "email", Message: "required"}})
```

## Status mapping

| Domain error | Status | `error` code |
|---|---|---|
| `ErrXxxNotFound` | 404 | `xxx_not_found` |
| `ErrXxxConflict` (e.g., email taken) | 409 | `xxx_conflict` |
| `ErrInvalidCredentials` | 401 | `invalid_credentials` |
| `ErrForbidden` | 403 | `forbidden` |
| Validation failure | 422 | `validation_failed` |
| Anything else | 500 | `internal` |

## Validation errors

When `validator.Struct(dto)` returns errors, convert to `[]FieldError{{Field, Message}}` — `Field` is JSON name, `Message` is human-readable. Surface via `response.ValidationErrors`.

## Never leak internals

```go
// WRONG
response.Error(w, 500, "internal", err.Error()) // raw err.Error() may say "ERROR: duplicate key on users_email_idx"
```

Use a constant string for the user-facing `message` and `log.Error().Err(err).Send()` for the detail.

## DO

```go
order, err := h.uc.Create(r.Context(), in)
switch {
case errors.Is(err, domain.ErrOrderEmpty):
    response.Error(w, http.StatusUnprocessableEntity, "empty_order", "order must have items")
case errors.Is(err, domain.ErrMenuItemUnavailable):
    response.Error(w, http.StatusConflict, "menu_item_unavailable", "one or more items are unavailable")
case err != nil:
    h.log.Err(err).Msg("create order failed")
    response.Error(w, http.StatusInternalServerError, "internal", "something went wrong")
default:
    response.OK(w, order)
}
```

## DON'T

```go
// no envelope
w.WriteHeader(200)
json.NewEncoder(w).Encode(order) // ← bypass response pkg

// status / code mismatch
response.Error(w, 200, "validation_failed", "…") // ← 2xx for an error
```
