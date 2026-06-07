# ADR-006 — YouTube Data API v3 for song search (API key only)

## Context
Song requests need a search backend. Options: YouTube, Spotify, last.fm. Spotify requires OAuth per user — incompatible with our anonymous customer flow. Last.fm has search but lacks streaming ID + thumbnail. YouTube via Data API v3 supports unauthenticated server-side API key calls and returns title, channelTitle, thumbnail, and videoId in one response.

## Decision
Use YouTube Data API v3 `search.list` endpoint, called server-side from `pkg/youtube`, with a single API key (`YOUTUBE_API_KEY`) loaded from env. The customer hits our `/api/customer/youtube/search?q=…` which the backend proxies to YouTube — we never expose the key to the browser.

## Consequences
- Pros: no OAuth complexity; no per-user state; cheap (10k units/day free; each search is 100 units → ~100 searches/day on the free tier, sufficient for a small cafe).
- Cons: cafe must obtain a Google Cloud API key during install. Documented in README quick start.
- Cost cap: if abuse appears, we add a Postgres-backed daily counter and gate the proxy at the per-day budget.
- Reversibility: `pkg/youtube` is a thin client; swapping to another provider is a 1-file change.
