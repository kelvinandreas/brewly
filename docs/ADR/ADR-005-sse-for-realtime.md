# ADR-005 — Server-Sent Events for real-time push

## Context
The KDS needs new orders to appear without polling. The song queue benefits from the same. Options: polling (simple, wasteful), SSE (one-way push, plain HTTP), WebSocket (bidirectional, needs upgrade handshake + framing).

## Decision
**Server-Sent Events**, served at `/api/sse/kitchen` and `/api/sse/song-queue`.

- Both feeds are one-way (server → client). We never push from client to server over them.
- Plain HTTP/1.1 long-lived connection; no protocol upgrade; works through standard nginx with `proxy_buffering off; proxy_read_timeout 1h;`.
- Browser API: `EventSource` — built in, auto-reconnects, no library needed.
- Backend: handler reads from a fan-out channel per topic, writes `data: <json>\n\n` per event, flushes after every write, sends `: keep-alive\n\n` every 15s.

## Consequences
- Pros: minimal moving parts; trivial to test (`curl -N http://localhost:8080/api/sse/kitchen` works); horizontal scaling needs sticky sessions or a pub/sub but we won't scale horizontally in v1.
- Cons: long-lived connections need careful goroutine lifecycle. We close on `r.Context().Done()` and `defer close(subscription)`. Documented in `sse-realtime` skill.
- Reversibility: client uses `EventSource` directly. Swapping to WebSocket later means replacing the hook (`useSseStream`) and the handler.
