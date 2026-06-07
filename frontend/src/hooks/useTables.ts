import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiFetch } from '../lib/api'
import type { Table, TableCreateResponse, TableRegenerateResponse } from '../types/api'

export const tableKeys = {
  all: ['tables'] as const,
}

export function useTables() {
  const queryClient = useQueryClient()

  const listQuery = useQuery({
    queryKey: tableKeys.all,
    queryFn: () =>
      apiFetch<{ tables: Table[] }>('/api/tables').then((d) => d.tables),
  })

  const createMutation = useMutation({
    mutationFn: (data: { label: string }) =>
      apiFetch<TableCreateResponse>('/api/tables', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: tableKeys.all }),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, label }: { id: string; label: string }) =>
      apiFetch<{ table: Table }>(`/api/tables/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ label }),
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: tableKeys.all }),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) =>
      apiFetch<void>(`/api/tables/${id}`, { method: 'DELETE' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: tableKeys.all }),
  })

  const regenerateTokenMutation = useMutation({
    mutationFn: (id: string) =>
      apiFetch<TableRegenerateResponse>(`/api/tables/${id}/regenerate-token`, {
        method: 'POST',
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: tableKeys.all }),
  })

  return { listQuery, createMutation, updateMutation, deleteMutation, regenerateTokenMutation }
}
