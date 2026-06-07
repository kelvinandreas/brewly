// Canonical TanStack Query hook pattern.
// Copy and rename Thing → your resource.
// Rules: all fetch calls go through lib/api.ts — never fetch directly in hooks.
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'

// Query key factory — export so components and prefetch can reuse exact keys.
export const thingKeys = {
  all: ['things'] as const,
  list: () => [...thingKeys.all, 'list'] as const,
  detail: (id: string) => [...thingKeys.all, 'detail', id] as const,
}

interface Thing {
  id: string
  name: string
  status: 'active' | 'archived'
}

interface CreateThingInput {
  name: string
}

// useThings — list query
export function useThings() {
  return useQuery({
    queryKey: thingKeys.list(),
    queryFn: () => api.get<Thing[]>('/api/things'),
  })
}

// useThing — single item query
export function useThing(id: string) {
  return useQuery({
    queryKey: thingKeys.detail(id),
    queryFn: () => api.get<Thing>(`/api/things/${id}`),
    enabled: !!id,
  })
}

// useCreateThing — mutation that invalidates the list on success
export function useCreateThing() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: CreateThingInput) => api.post<Thing>('/api/things', input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: thingKeys.list() })
    },
  })
}
