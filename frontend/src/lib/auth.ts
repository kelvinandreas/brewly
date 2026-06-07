// Module-scoped token store. Never written to localStorage/sessionStorage.
let _accessToken: string | null = null

export const getAccessToken = (): string | null => _accessToken
export const setAccessToken = (t: string): void => {
  _accessToken = t
}
export const clearAccessToken = (): void => {
  _accessToken = null
}

/** Silently exchange the httpOnly refresh cookie for a new access token. */
export async function refreshAccess(): Promise<string | null> {
  try {
    const res = await fetch('/api/auth/refresh', {
      method: 'POST',
      credentials: 'include',
    })
    if (!res.ok) {
      clearAccessToken()
      return null
    }
    const body = await res.json()
    setAccessToken(body.data.accessToken)
    return body.data.accessToken
  } catch {
    clearAccessToken()
    return null
  }
}
