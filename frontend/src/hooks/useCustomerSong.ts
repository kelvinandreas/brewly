import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getTableToken } from '../lib/tableAuth'
import type { SongRequest, YouTubeVideoResult } from '../types/api'

export const customerSongKeys = {
  search: (q: string) => ['customer-yt-search', q] as const,
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

export function useYouTubeSearch(q: string) {
  return useQuery({
    queryKey: customerSongKeys.search(q),
    queryFn: () =>
      customerFetch<{ results: YouTubeVideoResult[] }>(
        `/api/customer/songs/search?q=${encodeURIComponent(q)}&maxResults=10`,
      ).then((d) => d.results),
    enabled: q.trim().length >= 2,
    staleTime: 60_000,
  })
}

export function useSubmitSongRequest(tableId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: {
      videoId: string
      title: string
      channelName: string
      thumbnailUrl: string
      note?: string
    }) =>
      customerFetch<{ songRequest: SongRequest }>('/api/customer/songs', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ['customer-yt-search'] }),
    meta: { tableId },
  })
}
