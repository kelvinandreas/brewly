let _tableToken: string | null = null

export function setTableToken(token: string): void {
  _tableToken = token
}

export function getTableToken(): string | null {
  return _tableToken
}

export function clearTableToken(): void {
  _tableToken = null
}
