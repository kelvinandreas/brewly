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

export interface Category {
  id: string
  name: string
  displayOrder: number
  createdAt: string
  updatedAt: string
}

export interface MenuItem {
  id: string
  categoryId: string
  name: string
  description: string | null
  priceMinor: number
  imageUrl: string | null
  isAvailable: boolean
  createdAt: string
  updatedAt: string
}

export interface Table {
  id: string
  label: string
  tokenVersion: number
  createdAt: string
  updatedAt: string
}

export interface TableCreateResponse {
  table: Table
  qrToken: string
  qrUrl: string
}

export interface TableRegenerateResponse {
  qrToken: string
  qrUrl: string
}

export interface CustomerMenuItem {
  id: string
  name: string
  description: string | null
  priceMinor: number
  imageUrl: string | null
}

export interface CustomerMenuCategory {
  id: string
  name: string
  displayOrder: number
  items: CustomerMenuItem[]
}

export interface CustomerMenuResponse {
  categories: CustomerMenuCategory[]
}
