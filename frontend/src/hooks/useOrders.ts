import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiFetch } from '../lib/api'
import type { Order, PlaceOrderRequest } from '../types/api'

export const orderKeys = {
  all: ['orders'] as const,
  list: (filters: { status?: string; from?: string; to?: string }) =>
    [...orderKeys.all, filters] as const,
  detail: (id: string) => [...orderKeys.all, id] as const,
}

export function useOrders(filters: { status?: string; from?: string; to?: string } = {}) {
  const queryClient = useQueryClient()

  const params = new URLSearchParams()
  if (filters.status) params.set('status', filters.status)
  if (filters.from) params.set('from', filters.from)
  if (filters.to) params.set('to', filters.to)
  const qs = params.toString() ? `?${params.toString()}` : ''

  const listQuery = useQuery({
    queryKey: orderKeys.list(filters),
    queryFn: () => apiFetch<{ orders: Order[] }>(`/api/orders${qs}`).then((d) => d.orders),
  })

  const createMutation = useMutation({
    mutationFn: (data: { tableId: string } & PlaceOrderRequest) =>
      apiFetch<{ order: Order }>('/api/orders', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: orderKeys.all }),
  })

  const advanceStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      apiFetch<{ order: Order }>(`/api/orders/${id}/status`, {
        method: 'PATCH',
        body: JSON.stringify({ status }),
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: orderKeys.all }),
  })

  const cancelMutation = useMutation({
    mutationFn: (id: string) =>
      apiFetch<{ order: Order }>(`/api/orders/${id}/cancel`, { method: 'POST' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: orderKeys.all }),
  })

  return { listQuery, createMutation, advanceStatusMutation, cancelMutation }
}
