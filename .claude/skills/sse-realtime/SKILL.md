---
name: sse-realtime
description: Implementing SSE in Go (Chi) and consuming in React for Brewly's KDS and song queue. Covers handler headers, flush discipline, goroutine lifecycle, EventSource client, when SSE vs polling. Triggers when working on KDS, SSE endpoints, or real-time features.
---

## Server (Go)

```go
func (h *SSEHandler) Kitchen(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "streaming unsupported", http.StatusInternalServerError)
        return
    }

    sub := h.broker.Subscribe("kitchen")
    defer h.broker.Unsubscribe("kitchen", sub)

    keepAlive := time.NewTicker(15 * time.Second)
    defer keepAlive.Stop()

    for {
        select {
        case <-r.Context().Done():
            return
        case ev := <-sub:
            fmt.Fprintf(w, "event: %s\n", ev.Type)
            fmt.Fprintf(w, "data: %s\n\n", ev.JSON)
            flusher.Flush()
        case <-keepAlive.C:
            fmt.Fprintf(w, ": keep-alive\n\n")
            flusher.Flush()
        }
    }
}
```

## Broker pattern

```go
type Broker struct {
    mu   sync.RWMutex
    subs map[string]map[chan Event]struct{}
}

func (b *Broker) Subscribe(topic string) chan Event {
    ch := make(chan Event, 16) // buffered to absorb bursts
    b.mu.Lock()
    if b.subs[topic] == nil { b.subs[topic] = map[chan Event]struct{}{} }
    b.subs[topic][ch] = struct{}{}
    b.mu.Unlock()
    return ch
}

func (b *Broker) Unsubscribe(topic string, ch chan Event) {
    b.mu.Lock()
    delete(b.subs[topic], ch)
    b.mu.Unlock()
    close(ch)
}

func (b *Broker) Publish(topic string, ev Event) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    for ch := range b.subs[topic] {
        select { case ch <- ev: default: /* drop if slow consumer */ }
    }
}
```

Publishers (e.g., `OrderUsecase.Create`) call `b.Publish("kitchen", Event{...})` after successful commit.

## Client (React)

```tsx
// hooks/useSseStream.ts
export function useSseStream(path: string, onEvent: (e: MessageEvent) => void) {
  useEffect(() => {
    const es = new EventSource(`${import.meta.env.VITE_API_URL}${path}`, { withCredentials: true });
    es.addEventListener('order.created', onEvent);
    es.addEventListener('order.status_changed', onEvent);
    es.onerror = () => { /* EventSource auto-reconnects */ };
    return () => es.close();
  }, [path, onEvent]);
}
```

## When SSE vs polling

- **SSE** when updates are bursty + must be near-real-time AND you have a small number of clients (KDS: 1–5 kitchen tablets).
- **Polling** when updates are predictable + clients are many. Not applicable in Brewly v1.
- **WebSocket** when you need bidirectional. We don't.

## DO
- Always set `Connection: keep-alive` + `Cache-Control: no-cache`.
- Always check the response writer implements `http.Flusher` and `Flush()` after every write.
- Always exit on `r.Context().Done()` — closes the goroutine.
- Buffer subscription channels and drop on full (better than blocking the broker mutex).
- Configure nginx with `proxy_buffering off;` and `proxy_read_timeout 1h;`.

## DON'T
- Don't share a channel across topics — fan-out gets messy.
- Don't forget to `Unsubscribe` in `defer` — goroutine leak.
- Don't push events from within a DB transaction — push after commit (`tx.Commit()` succeeded).
- Don't write multi-line `data:` payloads — SSE requires `data: <oneline>\n` per line; either keep payload single-line JSON or send multiple `data:` lines.

## Test
- `curl -N http://localhost:8080/api/sse/kitchen -H "Authorization: Bearer …"` should print events as you POST orders in another shell.
