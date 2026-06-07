import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiFetch } from '../lib/api'
import type { Category } from '../types/api'

export const categoryKeys = {
  all: ['categories'] as const,
}

export function useCategories() {
  const queryClient = useQueryClient()

  const listQuery = useQuery({
    queryKey: categoryKeys.all,
    queryFn: () =>
      apiFetch<{ categories: Category[] }>('/api/categories').then((d) => d.categories),
  })

  const createMutation = useMutation({
    mutationFn: (data: { name: string; displayOrder: number }) =>
      apiFetch<{ category: Category }>('/api/categories', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: categoryKeys.all }),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, ...data }: { id: string; name?: string; displayOrder?: number }) =>
      apiFetch<{ category: Category }>(`/api/categories/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(data),
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: categoryKeys.all }),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) =>
      apiFetch<void>(`/api/categories/${id}`, { method: 'DELETE' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: categoryKeys.all }),
  })

  return { listQuery, createMutation, updateMutation, deleteMutation }
}
