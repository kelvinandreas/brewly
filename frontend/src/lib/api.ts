import type { ApiErrorResponse } from '../types/api'
import { clearAccessToken, getAccessToken, refreshAccess } from './auth'

export class ApiError extends Error {
  constructor(
    public readonly code: string,
    message: string,
    public readonly status: number,
    public readonly details?: ApiErrorResponse['details'],
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

async function parseError(res: Response): Promise<ApiError> {
  try {
    const body: ApiErrorResponse = await res.json()
    return new ApiError(body.error, body.message ?? res.statusText, res.status, body.details)
  } catch {
    return new ApiError('network_error', res.statusText, res.status)
  }
}

/** Fetch wrapper that injects the Bearer token and retries once on 401 after a silent refresh. */
export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const doFetch = async (token: string | null) => {
    const headers = new Headers(init.headers)
    headers.set('Content-Type', 'application/json')
    if (token) headers.set('Authorization', `Bearer ${token}`)
    return fetch(path, { ...init, headers, credentials: 'include' })
  }

  let res = await doFetch(getAccessToken())

  if (res.status === 401) {
    const newToken = await refreshAccess()
    if (!newToken) {
      clearAccessToken()
      throw new ApiError('unauthorized', 'Session expired', 401)
    }
    res = await doFetch(newToken)
  }

  if (!res.ok) throw await parseError(res)

  // 204 No Content
  if (res.status === 204) return undefined as T

  const body = await res.json()
  return body.data as T
}
