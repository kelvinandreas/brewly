# PRP — Songs & Reports (M4)

## Goal

Ship the two remaining v1 feature sets: the anonymous song-request flow (customer searches YouTube via backend proxy, submits a request, staff DJ board manages the queue live via SSE) and the owner reporting dashboard (daily/weekly/monthly revenue, best-selling items, hourly volume). When M4 ships, Brewly is feature-complete for v1.

## Context

- Schema per `memory-bank/database-schema.md`: `song_requests` table already defined in `001_init.sql`. No new migration needed.
- `song_requests.status` values: `queued → playing → played` or `queued → skipped`. Staff can only move forward; only one song can be `playing` at a time (enforced in usecase).
- `song_requests.token_jti` stores the JTI claim of the customer's table token. Used for rate limiting (max N requests per JTI per session). Per ADR-006.
- YouTube proxy per ADR-006: `GET /api/customer/youtube/search?q=` calls YouTube Data API v3 `search.list` from `pkg/youtube`. Key in `YOUTUBE_API_KEY` env. If key is empty, return `501 Not Implemented` with a clear message.
- Rate limit for song requests: max 3 active (`queued`) requests per `token_jti` at a time. Check in usecase against `song_requests.token_jti` count WHERE `status = 'queued'`. No Redis — pure Postgres count.
- SSE: `/api/sse/song-queue` uses same `pkg/sse.Broker` pattern as kitchen; token via `?token=` query param (same rationale as kitchen SSE). Events: `song.requested`, `song.status_changed`.
- Reports use raw SQL aggregations via GORM `db.Raw(...)` — per ADR-003 (GORM with raw fallback). No new domain entity needed; reports return lean DTOs directly from handler.
- Report date params: RFC3339 strings; server always interprets as UTC. `granularity` query param for revenue endpoint: `day`, `week`, `month`.
- Architecture per `memory-bank/architecture.md`: handler → usecase → domain ← repository. Report handler may call DB directly via a `ReportRepository` interface to avoid fat usecase for read-only aggregates.
- Frontend `/song-queue` route: staff DJ board showing queued + playing songs, advance status buttons, SSE-updated.
- Frontend `/reports` route: owner only; three chart-less but data-rich sections (revenue table, best-sellers list, hourly volume list) — no chart library in v1, plain tables.
- Customer song flow extends `/table/$tableId`: search field → YouTube results → "Request" button → active queue display.

## File structure

### To create

**Backend**
- `backend/internal/domain/song_request.go` — `SongRequest` struct + `SongRequestRepository` interface
- `backend/internal/domain/errors.go` additions — `ErrSongRequestNotFound`, `ErrSongRequestRateLimited`, `ErrInvalidSongStatusTransition`, `ErrSongAlreadyPlaying`
- `backend/internal/domain/constants.go` additions — song statuses (`SongQueued`, `SongPlaying`, `SongPlayed`, `SongSkipped`), `SongRequestRateLimit = 3`
- `backend/pkg/youtube/youtube.go` — `Client` struct, `NewClient(apiKey string)`, `Search(ctx, query string, maxResults int) ([]VideoResult, error)`. `VideoResult`: `{VideoID, Title, ChannelName, ThumbnailURL string}`. Returns `ErrKeyNotConfigured` if apiKey is empty.
- `backend/internal/repository/song_request_repo.go` — GORM impl: `Create`, `FindByID`, `List(ctx, status *string)`, `UpdateStatus`, `CountActiveByJTI(ctx, jti string) (int64, error)`
- `backend/internal/usecase/song_request.go` — `SongRequestUsecase`: `Submit`, `List`, `UpdateStatus`. `Submit` checks rate limit via `CountActiveByJTI`; validates status transitions.
- `backend/internal/usecase/song_request_test.go` — tests: submit succeeds, rate limit hit returns `ErrSongRequestRateLimited`, invalid transition returns `ErrInvalidSongStatusTransition`, two songs can't be playing simultaneously
- `backend/internal/usecase/report.go` — `ReportUsecase`: `Revenue(ctx, granularity, from, to)`, `BestSellers(ctx, from, to, limit)`, `HourlyVolume(ctx, date)`. Depends on `ReportRepository` interface.
- `backend/internal/domain/report.go` — lean DTOs + `ReportRepository` interface (raw SQL methods)
- `backend/internal/repository/report_repo.go` — GORM raw SQL impl
- `backend/internal/handler/song_request_handler.go` — staff endpoints: `GET /api/song-requests`, `PATCH /api/song-requests/:id/status`
- `backend/internal/handler/report_handler.go` — `GET /api/reports/revenue`, `GET /api/reports/best-sellers`, `GET /api/reports/hourly-volume`
- `backend/internal/handler/sse_handler.go` additions — `SongQueueStream` (GET /api/sse/song-queue)
- `frontend/src/hooks/useSongRequests.ts` — staff query + `updateStatusMutation`
- `frontend/src/hooks/useSongQueueSSE.ts` — `EventSource` for `/api/sse/song-queue`, reducer over `SongRequest[]`
- `frontend/src/hooks/useCustomerSong.ts` — YouTube search query + `submitSongMutation`
- `frontend/src/hooks/useReports.ts` — three separate queries (revenue, best-sellers, hourly-volume)
- `frontend/src/routes/_auth.song-queue.tsx` — DJ board: queued column + now playing, SSE-updated, advance/skip buttons
- `frontend/src/routes/_auth.reports.tsx` — owner reports page: revenue table + best-sellers + hourly volume

### To modify

- `backend/internal/domain/errors.go` — append song sentinels
- `backend/internal/domain/constants.go` — append song constants
- `backend/internal/handler/customer_handler.go` — add `YouTubeSearch` (GET /api/customer/youtube/search) and `SubmitSongRequest` (POST /api/customer/song-requests)
- `backend/internal/handler/sse_handler.go` — add `SongQueueStream`, add `songBroker *sse.Broker` field
- `backend/cmd/api/main.go` — wire song, report, new SSE routes; add `songBroker`, `YOUTUBE_API_KEY`
- `frontend/src/types/api.ts` — add `SongRequest`, `YouTubeVideoResult`, report DTOs
- `frontend/src/routes/table.$tableId.tsx` — add song search tab + queue display below menu
- `frontend/src/routes/_auth.dashboard.tsx` — add nav links to /song-queue and /reports
- `frontend/src/main.tsx` — register `songQueueRoute`, `reportsRoute`
- `memory-bank/progress.md` — tick M4 items

## Task breakdown

1. **Domain: SongRequest entity** — `domain/song_request.go`. `SongRequest` struct + `SongRequestRepository` interface with `Create`, `FindByID`, `List(ctx, status *string)`, `UpdateStatus`, `CountActiveByJTI`. Commit: `feat(domain): add SongRequest entity and repo interface`.

2. **Domain: report DTOs + interface** — `domain/report.go`. Lean structs: `RevenueRow{Period, TotalMinor, OrderCount}`, `BestSellerRow{MenuItemID, Name, TotalQty, TotalMinor}`, `HourlyVolumeRow{Hour int, OrderCount, TotalMinor}`. `ReportRepository` interface: `Revenue`, `BestSellers`, `HourlyVolume`. Commit: `feat(domain): add report DTOs and ReportRepository interface`.

3. **Domain: errors + constants** — append to `domain/errors.go` and `domain/constants.go`. Errors: `ErrSongRequestNotFound`, `ErrSongRequestRateLimited`, `ErrInvalidSongStatusTransition`, `ErrSongAlreadyPlaying`. Constants: `SongQueued/Playing/Played/Skipped`, `SongRequestRateLimit = 3`. Commit: `feat(domain): add song request constants and errors`.

4. **YouTube client** — `pkg/youtube/youtube.go`. `Client{apiKey}`, `NewClient(key)`, `Search(ctx, q, maxResults)` calls `https://www.googleapis.com/youtube/v3/search?part=snippet&type=video&q=…&maxResults=…&key=…`, parses `items[].snippet` into `[]VideoResult`. Returns `ErrKeyNotConfigured` sentinel when apiKey is empty. Commit: `feat(pkg): add YouTube search client`.

5. **Song request repository** — `repository/song_request_repo.go`. GORM impl; `CountActiveByJTI` counts rows WHERE `token_jti = ? AND status = 'queued'`. Commit: `feat(repo): add SongRequestRepo`.

6. **Report repository** — `repository/report_repo.go`. Three methods using `db.Raw(...)`:
   - `Revenue`: `SELECT date_trunc($granularity, created_at) AS period, SUM(total_minor), COUNT(*) FROM orders WHERE status='completed' AND created_at BETWEEN ? AND ? GROUP BY 1 ORDER BY 1`
   - `BestSellers`: `SELECT oi.menu_item_id, oi.name_snapshot AS name, SUM(oi.quantity), SUM(oi.quantity * oi.price_minor_snapshot) FROM order_items oi JOIN orders o ON o.id = oi.order_id WHERE o.status='completed' AND o.created_at BETWEEN ? AND ? GROUP BY 1, 2 ORDER BY 3 DESC LIMIT ?`
   - `HourlyVolume`: `SELECT EXTRACT(HOUR FROM created_at) AS hour, COUNT(*), SUM(total_minor) FROM orders WHERE status='completed' AND created_at::date = ? GROUP BY 1 ORDER BY 1`
   Commit: `feat(repo): add ReportRepo with raw SQL aggregations`.

7. **Song request usecase + tests** — `usecase/song_request.go`. `Submit(ctx, tableID, tokenJTI, videoID, title, channelName, thumbnailURL, note)`: checks `CountActiveByJTI < SongRequestRateLimit` else `ErrSongRequestRateLimited`; inserts with status `queued`; publishes `song.requested` SSE. `List(ctx, status *string)`. `UpdateStatus(ctx, id, newStatus)`: validates transition (only `queued→playing`, `playing→played`, `queued→skipped`, `playing→skipped`); if new status is `playing`, checks no other song is already `playing` else `ErrSongAlreadyPlaying`; publishes `song.status_changed`. Tests in `usecase/song_request_test.go`. Commit: `feat(usecase): add SongRequestUsecase with tests`.

8. **Report usecase** — `usecase/report.go`. Thin delegation to `ReportRepository`. `Revenue(ctx, granularity, from, to time.Time)`, `BestSellers(ctx, from, to time.Time, limit int)`, `HourlyVolume(ctx, date time.Time)`. Commit: `feat(usecase): add ReportUsecase`.

9. **SSEHandler: SongQueueStream** — extend `handler/sse_handler.go`. Add `songBroker *sse.Broker` field to `SSEHandler`; update `NewSSEHandler` signature. Add `SongQueueStream` GET `/api/sse/song-queue`: same pattern as `KitchenStream` — token query param → verify → subscribe → stream events + 15s keep-alive. Commit: `feat(handler): add SongQueueStream to SSEHandler`.

10. **Song request handler** — `handler/song_request_handler.go`. `List GET /api/song-requests?status=`, `UpdateStatus PATCH /api/song-requests/:id/status` with body `{status}`. Maps `ErrSongRequestNotFound → 404`, `ErrInvalidSongStatusTransition → 422`, `ErrSongAlreadyPlaying → 409`. Commit: `feat(handler): add SongRequestHandler`.

11. **Report handler** — `handler/report_handler.go`. Three endpoints; all owner-only. Parse `from`/`to` as RFC3339, `date` as `2006-01-02`, `granularity` as string, `limit` as int (default 10). Commit: `feat(handler): add ReportHandler`.

12. **CustomerHandler additions** — extend `handler/customer_handler.go`. `YouTubeSearch GET /api/customer/youtube/search?q=`: calls `youtubeClient.Search(ctx, q, 10)`, returns `{results: []}`. If `ErrKeyNotConfigured` return `501`. `SubmitSongRequest POST /api/customer/song-requests`: reads `tableID` and `tokenJTI` from ctx (injected by `RequireTableToken`); calls `songUC.Submit`. Commit: `feat(handler): add customer YouTube search and song request endpoints`.

13. **main.go wiring** — add `songBroker`, `youtubeClient`, `songRequestRepo`, `reportRepo`, `songRequestUC`, `reportUC`, `songRequestH`, `reportH`; update `NewSSEHandler` call; mount routes:
    - `/api/song-requests` (JWT cashier+owner)
    - `/api/sse/song-queue` (handler validates token)
    - `/api/reports/*` (JWT owner only)
    - `/api/customer/youtube/search` + `/api/customer/song-requests` (table token, inside existing group)
    Commit: `feat(api): wire song requests, reports, song-queue SSE in main.go`.

14. **`frontend/src/types/api.ts` additions** — `SongRequest`, `SongStatus`, `YouTubeVideoResult`, `RevenueRow`, `BestSellerRow`, `HourlyVolumeRow`. Commit: `feat(frontend): add SongRequest and report API types`.

15. **`frontend/src/hooks/useSongRequests.ts`** — query key `['song-requests', status]`, `updateStatusMutation`. Commit: `feat(frontend): add useSongRequests hook`.

16. **`frontend/src/hooks/useSongQueueSSE.ts`** — `EventSource('/api/sse/song-queue?token=…')`, reducer: add on `song.requested`, replace on `song.status_changed`. Commit: `feat(frontend): add useSongQueueSSE hook`.

17. **`frontend/src/hooks/useCustomerSong.ts`** — `useYouTubeSearch(q, enabled)` debounced query (500ms), `submitSongMutation`. Commit: `feat(frontend): add useCustomerSong hook`.

18. **`frontend/src/hooks/useReports.ts`** — three separate `useQuery` calls exported individually: `useRevenueReport(params)`, `useBestSellers(params)`, `useHourlyVolume(params)`. Commit: `feat(frontend): add useReports hooks`.

19. **`frontend/src/routes/_auth.song-queue.tsx`** — DJ board. "Now Playing" section at top (song with Playing status), "Queue" list below (queued songs), each card shows thumbnail, title, channel, requester table. Advance-status buttons: queued → "Play" or "Skip"; playing → "Mark Played" or "Skip". SSE-updated via `useSongQueueSSE`. Commit: `feat(frontend): add song queue DJ board route`.

20. **`frontend/src/routes/_auth.reports.tsx`** — owner reports. Date range picker (two `<input type="date">`), granularity selector. Three sections: Revenue table (period, orders, total), Best-sellers table (rank, name, qty, revenue), Hourly volume table (hour, orders, total). Commit: `feat(frontend): add reports route`.

21. **`frontend/src/routes/table.$tableId.tsx` update** — add "Request a Song" tab alongside menu. Tab switch between Menu and Songs. Songs tab: search `<input>` → debounced `useYouTubeSearch` → result grid (thumbnail + title + channel + Request button) → submitted songs list (active queue for this session). Commit: `feat(frontend): add song request flow to customer table route`.

22. **Wire routes + dashboard** — update `main.tsx` (add `songQueueRoute`, `reportsRoute`), update `_auth.dashboard.tsx` (add nav links). Commit: `feat(frontend): wire song queue and reports routes`.

23. **Memory bank** — tick M4 items in `progress.md`. Commit: `chore(docs): update memory-bank after M4 songs, reports`.

## Pseudocode (trickiest: UpdateStatus with playing-lock)

```go
// usecase/song_request.go

// validSongTransitions maps current → allowed next statuses (set).
var validSongTransitions = map[string]map[string]bool{
    domain.SongQueued:  {domain.SongPlaying: true, domain.SongSkipped: true},
    domain.SongPlaying: {domain.SongPlayed: true,  domain.SongSkipped: true},
}

func (u *SongRequestUsecase) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus string) (*domain.SongRequest, error) {
    sr, err := u.repo.FindByID(ctx, id)
    if err != nil { return nil, fmt.Errorf("usecase.SongRequest.UpdateStatus: %w", err) }

    allowed, ok := validSongTransitions[sr.Status]
    if !ok || !allowed[newStatus] {
        return nil, domain.ErrInvalidSongStatusTransition
    }

    // Enforce single-playing invariant.
    if newStatus == domain.SongPlaying {
        playing, err := u.repo.CountByStatus(ctx, domain.SongPlaying)
        if err != nil { return nil, fmt.Errorf("usecase.SongRequest.UpdateStatus: %w", err) }
        if playing > 0 { return nil, domain.ErrSongAlreadyPlaying }
    }

    if err := u.repo.UpdateStatus(ctx, id, newStatus); err != nil {
        return nil, fmt.Errorf("usecase.SongRequest.UpdateStatus: %w", err)
    }
    sr.Status = newStatus
    payload, _ := json.Marshal(sr)
    u.broker.Publish(sse.Event{Type: "song.status_changed", Payload: payload})
    return sr, nil
}
```

## Validation plan

- `make lint` clean.
- `make test` — all existing tests still pass; new `usecase/song_request_test.go` covers: submit ok, submit at rate limit fails, invalid transition fails, double-playing fails.
- Manual smoke (with `make dev`):
  1. Scan QR → Songs tab → search "jazz" → results appear → click Request.
  2. Owner opens `/song-queue` → song appears in queue.
  3. DJ clicks Play → song moves to Now Playing.
  4. DJ clicks Mark Played → song leaves Now Playing.
  5. Request 4 songs from same session → 4th returns 429 (rate limited).
  6. Owner opens `/reports` → revenue table populates after completing a few orders.
  7. `curl -N "http://localhost:8080/api/sse/song-queue?token=…"` → streams keep-alive.
  8. YouTube search with no `YOUTUBE_API_KEY` set → 501 response.

## Out of scope

- Song request deduplication (same video can be requested multiple times in v1).
- Per-table song limit (only per-JTI limit in v1).
- Chart rendering (plain tables in v1; chart library is a future enhancement).
- Revenue export (CSV / PDF — deferred).
- Pagination on reports (small cafe, short date ranges assumed).
