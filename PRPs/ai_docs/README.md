# AI docs cache

Snapshots of upstream documentation Sonnet should fetch and save here so PRPs can cite stable references without re-fetching mid-task. Add a dated subfolder per fetch.

Required initial snapshots (fetch on M0):

- `tanstack-router/` — file-based routing reference, route loader pattern
- `tanstack-query/` — useQuery, useMutation, query keys, invalidation
- `gorm/` — model definitions, soft delete, hooks, raw SQL
- `chi/` — router, sub-routers, middleware signature
- `youtube-data-api/` — `search.list` reference (parts, quota costs)
- `golang-jwt-v5/` — Sign, Parse, Claims

Use `WebFetch` via Claude to grab the relevant pages; commit the markdown copies.
