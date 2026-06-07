# Progress

Legend: `[ ]` todo · `[~]` in progress · `[x]` done

## M0 — Bootstrap

- [x] Root MD files (CLAUDE.md, AGENTS.md, README.md, CONTRIBUTING.md, LICENSE)
- [x] memory-bank/ files
- [x] docs/ENGINEERING.md + docs/specification.md
- [x] All 6 ADRs
- [x] .claude/commands, .claude/rules, .claude/skills
- [x] PRPs/ templates + auth example
- [x] examples/ canonical files
- [x] backend/ go.mod, main.go boots and responds 200 to `/healthz`
- [x] frontend/ pnpm install, `vite dev` serves the React shell
- [x] docker-compose.yml `make dev` brings up pg + backend + frontend cleanly
- [x] Makefile targets all execute

## M1 — Auth

- [x] migrations/001_init.sql committed (full schema — all 8 tables)
- [x] migrations/002_seed_owner.sql (placeholder; registration is self-serve via API)
- [x] backend: domain.User, repository.UserRepo, usecase.AuthUsecase, handler.AuthHandler
- [x] middleware.JWTAuth (RequireAuth with role enforcement)
- [x] frontend: /login route, useAuth hook, memory token store
- [x] update memory-bank/api-contracts.md
- [x] update memory-bank/database-schema.md

## M2 — Menu, tables, QR

- [x] backend: domain.Category, domain.MenuItem, domain.Table (entities + repo interfaces)
- [x] backend: domain/errors.go additions (ErrCategoryNotFound, ErrMenuItemNotFound, ErrTableNotFound, ErrTableLabelTaken, ErrMenuItemUnavailable)
- [x] backend: domain/constants.go additions (ContextKeyTableID, ContextKeyTokenJTI)
- [x] backend: pkg/tabletoken — Sign/Verify, 4h TTL, claims {tid, tvr, jti}
- [x] backend: pkg/qrcode — Generate() returns 256×256 PNG bytes
- [x] backend: repository.CategoryRepo, repository.MenuItemRepo, repository.TableRepo (GORM)
- [x] backend: middleware.RequireTableToken (sig → exp → DB lookup → token_version check → ctx inject)
- [x] backend: usecase.CategoryUsecase, usecase.MenuItemUsecase, usecase.TableUsecase
- [x] backend: handler.CategoryHandler, handler.MenuItemHandler, handler.TableHandler, handler.CustomerHandler
- [x] backend: main.go wired (all M2 routes under correct auth middleware)
- [x] frontend: types/api.ts additions (Category, MenuItem, Table, QR types, CustomerMenu types)
- [x] frontend: lib/tableAuth.ts (module-scope table token store)
- [x] frontend: hooks/useCategories, hooks/useMenuItems, hooks/useTables
- [x] frontend: routes/_auth.menu.tsx (owner menu management, category + item CRUD)
- [x] frontend: routes/_auth.tables.tsx (table list, QR modal, regen token)
- [x] frontend: routes/table.$tableId.tsx (customer QR landing + menu browse)
- [x] frontend: main.tsx route wiring; dashboard nav links updated

## M3 — Orders, payments, KDS

- [x] backend: domain.Order, domain.OrderItem, domain.Payment (entities + repo interfaces)
- [x] backend: domain/errors.go additions (ErrOrderNotFound, ErrOrderCancelled, ErrInvalidStatusTransition, ErrPaymentConflict)
- [x] backend: domain/constants.go additions (order statuses, order sources, payment methods)
- [x] backend: pkg/sse — generic fan-out SSE Broker
- [x] backend: repository.OrderRepo (Create with items in tx, List, FindByID, UpdateStatus, ListByTable)
- [x] backend: repository.PaymentRepo
- [x] backend: usecase.OrderUsecase (state machine, SSE publish, tests)
- [x] backend: usecase.PaymentUsecase
- [x] backend: handler.SSEHandler — GET /api/sse/kitchen (token via query param)
- [x] backend: handler.OrderHandler — staff CRUD + status + cancel
- [x] backend: handler.PaymentHandler
- [x] backend: handler.CustomerHandler additions — POST /api/customer/orders, GET /api/customer/orders/mine
- [x] backend: main.go wired (all M3 routes, SSE broker)
- [x] frontend: types/api.ts additions (Order, OrderItem, Payment, KitchenSSEEvent)
- [x] frontend: hooks/useOrders, hooks/useCustomerOrder, hooks/useKitchenSSE
- [x] frontend: routes/_auth.kitchen.tsx (KDS board, SSE-powered, 4-column status layout)
- [x] frontend: routes/_auth.cashier.tsx (Cashier POS — table picker, menu, cart, payment form)
- [x] frontend: routes/_auth.orders.tsx (order history with status filter tabs)
- [x] frontend: routes/table.$tableId.tsx updated (cart + order placement + my orders)
- [x] frontend: main.tsx + dashboard nav links updated

## M4 — Songs, reports

- [x] backend: domain.SongRequest (entity + repo interface); domain.RevenueRow, BestSellerRow, HourlyVolumeRow; domain.ReportRepository
- [x] backend: domain/errors.go additions (ErrSongRequestNotFound, ErrSongRequestRateLimited, ErrInvalidSongStatusTransition, ErrSongAlreadyPlaying)
- [x] backend: domain/constants.go additions (SongQueued/Playing/Played/Skipped, SongRequestRateLimit=3)
- [x] backend: pkg/youtube — YouTube Data API v3 client; ErrKeyNotConfigured sentinel
- [x] backend: repository.SongRequestRepo (GORM, CountActiveByJTI, CountByStatus)
- [x] backend: repository.ReportRepo (3 raw SQL aggregations: Revenue/BestSellers/HourlyVolume)
- [x] backend: usecase.SongRequestUsecase (state machine, rate limit, single-playing invariant, SSE publish, tests)
- [x] backend: usecase.ReportUsecase (thin delegation)
- [x] backend: handler.SSEHandler updated — added songBroker + SongQueueStream; refactored shared stream() method
- [x] backend: handler.SongRequestHandler — GET /api/song-requests, PATCH /api/song-requests/:id/status
- [x] backend: handler.ReportHandler — GET /api/reports/revenue, /best-sellers, /hourly-volume
- [x] backend: handler.CustomerHandler additions — GET /api/customer/songs/search, POST /api/customer/songs
- [x] backend: main.go wired (all M4 routes: song-requests, reports, SSE song-queue, customer songs)
- [x] frontend: types/api.ts additions (SongRequest, SongStatus, YouTubeVideoResult, SongQueueSSEEvent, RevenueRow, BestSellerRow, HourlyVolumeRow)
- [x] frontend: hooks/useSongRequests (staff query + updateStatus mutation)
- [x] frontend: hooks/useSongQueueSSE (useReducer with SSE fan-in for song queue)
- [x] frontend: hooks/useCustomerSong (useYouTubeSearch + useSubmitSongRequest)
- [x] frontend: hooks/useReports (useRevenueReport, useBestSellersReport, useHourlyVolumeReport)
- [x] frontend: routes/_auth.song-queue.tsx (DJ board — 4-column, SSE-powered, skip/play actions)
- [x] frontend: routes/_auth.reports.tsx (revenue + best-sellers + hourly tabs with date pickers)
- [x] frontend: routes/table.$tableId.tsx updated (Songs tab — YouTube search, debounced, request with note)
- [x] frontend: main.tsx + dashboard nav links updated
