---
name: tanstack-patterns
description: TanStack Router/Query conventions for Brewly — useQuery/useMutation structure, route loader pattern, query key conventions, cache invalidation, optimistic updates. Triggers when editing tsx files in routes/, hooks/, or components/features/.
---

## Query keys

Top of every `useXxx.ts` exports a `xxxKeys` object:

```ts
export const orderKeys = {
  all: ['orders'] as const,
  lists: () => [...orderKeys.all, 'list'] as const,
  list: (filters: OrderFilters) => [...orderKeys.lists(), filters] as const,
  detail: (id: string) => [...orderKeys.all, 'detail', id] as const,
};
```

Why: every cached entry sits under `['orders']`, so `qc.invalidateQueries({ queryKey: orderKeys.all })` after a mutation clears everything order-related at once.

## Hook shape

```ts
// useOrders.ts
export const useOrders = (filters: OrderFilters) =>
  useQuery({
    queryKey: orderKeys.list(filters),
    queryFn: ({ signal }) => api.get<Order[]>('/orders', { params: filters, signal }),
    staleTime: 5_000,
  });

export const useCreateOrder = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: NewOrder) => api.post<Order>('/orders', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: orderKeys.lists() }),
  });
};
```

## Route loaders

Use loaders for first-paint data; rely on `useQuery` for subsequent updates. The loader and the hook share the query function via `queryClient.ensureQueryData`.

```ts
export const Route = createFileRoute('/_auth/orders/')({
  loader: ({ context: { queryClient } }) =>
    queryClient.ensureQueryData({
      queryKey: orderKeys.list({}),
      queryFn: () => api.get<Order[]>('/orders'),
    }),
  component: OrdersPage,
});
```

## Optimistic updates (status changes)

```ts
useMutation({
  mutationFn: ({ id, status }) => api.patch(`/orders/${id}/status`, { status }),
  onMutate: async ({ id, status }) => {
    await qc.cancelQueries({ queryKey: orderKeys.detail(id) });
    const prev = qc.getQueryData<Order>(orderKeys.detail(id));
    qc.setQueryData<Order>(orderKeys.detail(id), (o) => o && { ...o, status });
    return { prev };
  },
  onError: (_e, { id }, ctx) => ctx?.prev && qc.setQueryData(orderKeys.detail(id), ctx.prev),
  onSettled: ({ id }) => qc.invalidateQueries({ queryKey: orderKeys.detail(id) }),
});
```

## DO

- Pass `signal` from `queryFn` arg into `api.*` for cancellation.
- Co-locate keys with the hook that uses them.
- Invalidate at the smallest level that's still correct (`orderKeys.lists()` for new orders; the specific `detail()` for status edits).

## DON'T

```ts
// No magic strings as keys
useQuery({ queryKey: ['orders', 'list', filters], queryFn: ... }); // ← drift across files

// No fetch in components
function OrdersPage() {
  useEffect(() => { fetch('/api/orders').then(...) }, []); // ← move to a hook
}

// No cross-domain invalidation
qc.invalidateQueries(); // ← nukes everything; pick a key
```
