import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { LoginRequest, RegisterOwnerRequest, TokenResponse, User } from '../types/api'
import { apiFetch, ApiError } from '../lib/api'
import { clearAccessToken, refreshAccess, setAccessToken } from '../lib/auth'

export const authKeys = {
  me: ['auth', 'me'] as const,
}

export function useAuth() {
  const queryClient = useQueryClient()

  const meQuery = useQuery({
    queryKey: authKeys.me,
    queryFn: () => apiFetch<{ user: User }>('/api/auth/me').then((d) => d.user),
    retry: false,
    staleTime: 5 * 60 * 1000,
  })

  const loginMutation = useMutation({
    mutationFn: (data: LoginRequest) =>
      apiFetch<TokenResponse>('/api/auth/login', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: (data) => {
      if (data.accessToken) setAccessToken(data.accessToken)
      queryClient.invalidateQueries({ queryKey: authKeys.me })
    },
  })

  const registerOwnerMutation = useMutation({
    mutationFn: (data: RegisterOwnerRequest) =>
      apiFetch<TokenResponse>('/api/auth/register-owner', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: authKeys.me })
    },
  })

  const logoutMutation = useMutation({
    mutationFn: () => apiFetch('/api/auth/logout', { method: 'POST' }),
    onSettled: () => {
      clearAccessToken()
      queryClient.clear()
      // Full page reset on logout — clears all in-memory state cleanly.
      window.location.assign('/login')
    },
  })

  /** Try to get a valid access token, refreshing silently if needed. Returns null if unauthenticated. */
  async function ensureAccessToken(): Promise<string | null> {
    const { getAccessToken } = await import('../lib/auth')
    const existing = getAccessToken()
    if (existing) return existing
    return refreshAccess()
  }

  return {
    user: meQuery.data ?? null,
    isLoading: meQuery.isLoading,
    isOwnerNotExists: meQuery.error instanceof ApiError && meQuery.error.status === 404,
    login: loginMutation,
    registerOwner: registerOwnerMutation,
    logout: logoutMutation,
    ensureAccessToken,
  }
}
