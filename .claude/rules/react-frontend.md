# React frontend rules

Globs: `**/*.tsx`, `**/*.ts`

## Components
- One component per file. Named + default export of the same identifier.
- No business logic in components — only display + handler wiring.
- All data fetching/mutation through a TanStack Query hook in `src/hooks/`.

## Hooks
- Query key is a tuple constant exported at the top of the hook file: `export const orderKeys = { all: ['orders'] as const, list: (filters) => [...orderKeys.all, filters] as const };`.
- Mutations invalidate the parent list key on success.

## Routes
- File-based. Protected routes nest under `_auth.tsx` whose `beforeLoad` redirects to `/login` if no access token.
- Customer route `table/$tableId.tsx` reads the token from URL on mount, stores in memory, replaces URL.

## Styling
- Tailwind utility-first. Component-level CSS only when escape hatch needed.
- All currency display via `formatIDR()` from `src/lib/currency.ts`.

## Forms
- React Hook Form + Zod resolver. Schema in the same file as the form.

## State
- Server state → TanStack Query. UI state → component state or a focused context.
- **Never** `localStorage` / `sessionStorage` for tokens.

## TS
- `strict: true`. Avoid `any`; if necessary, isolate to one line + comment.
- Backend types live in `src/types/api.ts` and mirror Go DTOs.
