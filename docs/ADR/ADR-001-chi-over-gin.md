# ADR-001 — Use Chi as the HTTP router

## Context
The backend needs an HTTP router. The popular Go options are net/http alone, Gin, Echo, Fiber, and Chi. The codebase favors clean architecture where handlers are thin and the router stays unopinionated; we also want middleware composition to feel like wrapping `http.Handler` (so generic middleware libs work) rather than a framework-specific signature.

## Decision
Use `github.com/go-chi/chi/v5`.

- Idiomatic: `chi.Mux` implements `http.Handler`; middleware is `func(http.Handler) http.Handler`.
- Lightweight: small dependency tree, no reflection at request time.
- First-class sub-routers via `chi.Route`, which maps cleanly to our domain grouping.
- Existing middleware ecosystem (`go-chi/cors`, `go-chi/jwtauth`) we may pull selectively.

Rejected alternatives:
- **Gin**: faster benchmarks but its `gin.Context` type bleeds into handlers, complicating the handler/usecase boundary and making middleware non-portable.
- **Echo / Fiber**: same coupling problem; Fiber's fasthttp also rules out using stdlib `net/http` features cleanly.
- **net/http alone**: routing/params boilerplate doesn't earn its keep at our endpoint count (~40).

## Consequences
- Pros: clean architecture stays clean; middleware portable; learning curve is essentially stdlib + thin router.
- Cons: slightly more verbose route declarations than Gin; have to assemble JSON helpers ourselves (we use `pkg/response`).
- Reversibility: handlers stay framework-agnostic by reading from `*http.Request`, so switching routers later is a Friday afternoon.
