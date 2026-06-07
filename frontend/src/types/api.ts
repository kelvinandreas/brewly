// API types — mirror Go DTOs exactly.

export interface User {
  id: string
  email: string
  name: string
  role: 'owner' | 'cashier' | 'kitchen'
  createdAt: string
  updatedAt: string
}

export interface ApiResponse<T> {
  success: true
  data: T
  message?: string
}

export interface ApiErrorResponse {
  success: false
  error: string
  message?: string
  details?: FieldError[]
}

export interface FieldError {
  field: string
  message: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterOwnerRequest {
  email: string
  password: string
  name: string
}

export interface TokenResponse {
  accessToken: string
  user?: User
}
