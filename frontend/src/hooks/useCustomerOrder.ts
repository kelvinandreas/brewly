import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getTableToken } from '../lib/tableAuth'
import type { Order, PlaceOrderRequest } from '../types/api'

export const customerOrderKeys = {
  mine: (tableId: string) => ['customer-orders', tableId] as const,
}

async function customerFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getTableToken()
  const res = await fetch(path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  })
  const body = await res.json()
  if (!res.ok) throw new Error(body?.error ?? 'Request failed')
  return body.data as T
}

export function useCustomerOrder(tableId: string) {
  const queryClient = useQueryClient()

  const myOrdersQuery = useQuery({
    queryKey: customerOrderKeys.mine(tableId),
    queryFn: () =>
      customerFetch<{ orders: Order[] }>('/api/customer/orders/mine').then((d) => d.orders),
    enabled: !!getTableToken(),
  })

  const placeOrderMutation = useMutation({
    mutationFn: (data: PlaceOrderRequest) =>
      customerFetch<{ order: Order }>('/api/customer/orders', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: customerOrderKeys.mine(tableId) }),
  })

  return { myOrdersQuery, placeOrderMutation }
}
