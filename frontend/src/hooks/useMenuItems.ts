import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiFetch } from '../lib/api'
import type { MenuItem } from '../types/api'

export const menuItemKeys = {
  all: ['menu-items'] as const,
  list: (filters: { categoryId?: string; availableOnly?: boolean }) =>
    [...menuItemKeys.all, filters] as const,
}

export function useMenuItems(filters: { categoryId?: string; availableOnly?: boolean } = {}) {
  const queryClient = useQueryClient()

  const params = new URLSearchParams()
  if (filters.categoryId) params.set('categoryId', filters.categoryId)
  if (filters.availableOnly) params.set('availableOnly', 'true')
  const qs = params.toString() ? `?${params.toString()}` : ''

  const listQuery = useQuery({
    queryKey: menuItemKeys.list(filters),
    queryFn: () =>
      apiFetch<{ items: MenuItem[] }>(`/api/menu-items${qs}`).then((d) => d.items),
  })

  const createMutation = useMutation({
    mutationFn: (data: {
      categoryId: string
      name: string
      description?: string | null
      priceMinor: number
      imageUrl?: string | null
      isAvailable?: boolean
    }) =>
      apiFetch<{ item: MenuItem }>('/api/menu-items', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: menuItemKeys.all }),
  })

  const updateMutation = useMutation({
    mutationFn: ({
      id,
      ...data
    }: {
      id: string
      categoryId?: string
      name?: string
      description?: string | null
      priceMinor?: number
      imageUrl?: string | null
      isAvailable?: boolean
    }) =>
      apiFetch<{ item: MenuItem }>(`/api/menu-items/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(data),
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: menuItemKeys.all }),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) =>
      apiFetch<void>(`/api/menu-items/${id}`, { method: 'DELETE' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: menuItemKeys.all }),
  })

  return { listQuery, createMutation, updateMutation, deleteMutation }
}
