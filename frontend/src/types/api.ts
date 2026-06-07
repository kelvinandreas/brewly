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

export interface OrderItem {
  id: string
  orderId: string
  menuItemId: string
  nameSnapshot: string
  priceMinorSnapshot: number
  quantity: number
  createdAt: string
  updatedAt: string
}

export type OrderStatus = 'pending' | 'confirmed' | 'preparing' | 'ready' | 'completed' | 'cancelled'
export type OrderSource = 'customer_qr' | 'cashier_pos'

export interface Order {
  id: string
  tableId: string
  status: OrderStatus
  source: OrderSource
  totalMinor: number
  note: string
  createdByUserId: string | null
  items: OrderItem[]
  createdAt: string
  updatedAt: string
}

export type PaymentMethod = 'cash' | 'qris' | 'card'

export interface Payment {
  id: string
  orderId: string
  method: PaymentMethod
  amountMinor: number
  receivedMinor: number
  recordedByUserId: string
  createdAt: string
}

export interface PlaceOrderRequest {
  items: { menuItemId: string; quantity: number }[]
  note?: string
}

export interface KitchenSSEEvent {
  type: 'order.created' | 'order.status_changed' | 'order.cancelled'
  payload: Order
}

export type SongStatus = 'queued' | 'playing' | 'played' | 'skipped'

export interface SongRequest {
  id: string
  tableId: string
  tokenJti: string
  youtubeVideoId: string
  title: string
  channelName: string
  thumbnailUrl: string
  note: string
  status: SongStatus
  createdAt: string
  updatedAt: string
}

export interface YouTubeVideoResult {
  videoId: string
  title: string
  channelName: string
  thumbnailUrl: string
}

export interface SongQueueSSEEvent {
  type: 'song.requested' | 'song.status_changed'
  payload: SongRequest
}

export interface RevenueRow {
  period: string
  totalMinor: number
  orderCount: number
}

export interface BestSellerRow {
  menuItemId: string
  name: string
  totalQuantity: number
  totalMinor: number
}

export interface HourlyVolumeRow {
  hour: number
  orderCount: number
  totalMinor: number
}
