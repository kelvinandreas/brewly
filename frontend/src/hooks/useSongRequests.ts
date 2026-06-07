import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiFetch } from '../lib/api'
import type { SongRequest, SongStatus } from '../types/api'

export const songRequestKeys = {
  all: ['song-requests'] as const,
  list: (status?: SongStatus) => [...songRequestKeys.all, { status }] as const,
}

export function useSongRequests(status?: SongStatus) {
  const queryClient = useQueryClient()

  const params = new URLSearchParams()
  if (status) params.set('status', status)
  const qs = params.toString() ? `?${params.toString()}` : ''

  const listQuery = useQuery({
    queryKey: songRequestKeys.list(status),
    queryFn: () =>
      apiFetch<{ songRequests: SongRequest[] }>(`/api/song-requests${qs}`).then(
        (d) => d.songRequests,
      ),
  })

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status: newStatus }: { id: string; status: SongStatus }) =>
      apiFetch<{ songRequest: SongRequest }>(`/api/song-requests/${id}/status`, {
        method: 'PATCH',
        body: JSON.stringify({ status: newStatus }),
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: songRequestKeys.all }),
  })

  return { listQuery, updateStatusMutation }
}
