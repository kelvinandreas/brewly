import { createRoute } from '@tanstack/react-router'
import { Route as authRoute } from './_auth'
import { useOrders } from '../hooks/useOrders'
import { useKitchenSSE } from '../hooks/useKitchenSSE'
import { getAccessToken } from '../lib/auth'
import type { Order, OrderStatus } from '../types/api'
import { formatIDR } from '../lib/currency'

export const Route = createRoute({
  getParentRoute: () => authRoute,
  path: '/kitchen',
  component: KitchenPage,
})

const COLUMNS: { status: OrderStatus; label: string; color: string }[] = [
  { status: 'pending', label: 'Pending', color: 'border-yellow-400 bg-yellow-50' },
  { status: 'confirmed', label: 'Confirmed', color: 'border-blue-400 bg-blue-50' },
  { status: 'preparing', label: 'Preparing', color: 'border-orange-400 bg-orange-50' },
  { status: 'ready', label: 'Ready', color: 'border-green-400 bg-green-50' },
]

const NEXT_STATUS: Partial<Record<OrderStatus, OrderStatus>> = {
  pending: 'confirmed',
  confirmed: 'preparing',
  preparing: 'ready',
  ready: 'completed',
}

const NEXT_LABEL: Partial<Record<OrderStatus, string>> = {
  pending: 'Confirm',
  confirmed: 'Start',
  preparing: 'Mark Ready',
  ready: 'Complete',
}

function elapsed(createdAt: string): string {
  const secs = Math.floor((Date.now() - new Date(createdAt).getTime()) / 1000)
  if (secs < 60) return `${secs}s`
  const mins = Math.floor(secs / 60)
  return `${mins}m`
}

function KitchenPage() {
  const { listQuery, advanceStatusMutation, cancelMutation } = useOrders()
  const initialOrders = listQuery.data ?? []
  const orders = useKitchenSSE(getAccessToken(), initialOrders)

  const activeOrders = orders.filter((o) => o.status !== 'completed' && o.status !== 'cancelled')

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <header className="bg-gray-800 px-6 py-4 flex items-center justify-between">
        <h1 className="text-xl font-bold text-amber-400">Kitchen Display</h1>
        <span className="text-xs text-gray-400">{activeOrders.length} active</span>
      </header>

      <div className="p-4 grid grid-cols-2 lg:grid-cols-4 gap-4 items-start">
        {COLUMNS.map((col) => {
          const colOrders = activeOrders.filter((o) => o.status === col.status)
          return (
            <div key={col.status}>
              <div className="flex items-center gap-2 mb-3">
                <h2 className="text-sm font-semibold uppercase tracking-wide text-gray-300">
                  {col.label}
                </h2>
                <span className="bg-gray-700 text-gray-300 text-xs px-1.5 py-0.5 rounded-full">
                  {colOrders.length}
                </span>
              </div>
              <div className="space-y-3">
                {colOrders.map((order) => (
                  <OrderCard
                    key={order.id}
                    order={order}
                    colorClass={col.color}
                    onAdvance={() => {
                      const next = NEXT_STATUS[order.status as OrderStatus]
                      if (next) advanceStatusMutation.mutate({ id: order.id, status: next })
                    }}
                    nextLabel={NEXT_LABEL[order.status as OrderStatus]}
                    onCancel={() => {
                      if (confirm('Cancel this order?')) cancelMutation.mutate(order.id)
                    }}
                  />
                ))}
                {colOrders.length === 0 && (
                  <p className="text-xs text-gray-600 text-center py-4">Empty</p>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function OrderCard({
  order,
  colorClass,
  onAdvance,
  nextLabel,
  onCancel,
}: {
  order: Order
  colorClass: string
  onAdvance: () => void
  nextLabel?: string
  onCancel: () => void
}) {
  return (
    <div className={`rounded-lg border-2 ${colorClass} p-3 text-gray-800`}>
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-bold uppercase text-gray-500">
          {order.source === 'customer_qr' ? `Table ${order.tableId.slice(-4)}` : 'Counter'}
        </span>
        <span className="text-xs text-gray-400">{elapsed(order.createdAt)} ago</span>
      </div>
      <ul className="space-y-1 mb-3">
        {order.items.map((item) => (
          <li key={item.id} className="flex justify-between text-sm">
            <span className="font-medium">{item.nameSnapshot}</span>
            <span className="text-gray-500">×{item.quantity}</span>
          </li>
        ))}
      </ul>
      {order.note && <p className="text-xs text-gray-500 italic mb-2">{order.note}</p>}
      <div className="flex items-center justify-between">
        <span className="text-xs font-bold text-amber-600">{formatIDR(order.totalMinor)}</span>
        <div className="flex gap-1">
          <button
            onClick={onCancel}
            className="text-xs text-red-500 hover:text-red-700 px-1.5 py-1 rounded"
          >
            ✕
          </button>
          {nextLabel && (
            <button
              onClick={onAdvance}
              className="text-xs bg-gray-800 text-white px-2 py-1 rounded hover:bg-gray-700"
            >
              {nextLabel}
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
