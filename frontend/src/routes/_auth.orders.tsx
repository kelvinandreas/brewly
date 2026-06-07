import { createRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { Route as authRoute } from './_auth'
import { useOrders } from '../hooks/useOrders'
import type { OrderStatus } from '../types/api'
import { formatIDR } from '../lib/currency'

export const Route = createRoute({
  getParentRoute: () => authRoute,
  path: '/orders',
  component: OrdersPage,
})

const STATUS_TABS: { value: string; label: string }[] = [
  { value: '', label: 'All' },
  { value: 'pending', label: 'Pending' },
  { value: 'confirmed', label: 'Confirmed' },
  { value: 'preparing', label: 'Preparing' },
  { value: 'ready', label: 'Ready' },
  { value: 'completed', label: 'Completed' },
  { value: 'cancelled', label: 'Cancelled' },
]

const STATUS_COLORS: Record<OrderStatus, string> = {
  pending: 'bg-yellow-100 text-yellow-700',
  confirmed: 'bg-blue-100 text-blue-700',
  preparing: 'bg-orange-100 text-orange-700',
  ready: 'bg-green-100 text-green-700',
  completed: 'bg-gray-100 text-gray-600',
  cancelled: 'bg-red-100 text-red-600',
}

function OrdersPage() {
  const [statusFilter, setStatusFilter] = useState('')
  const { listQuery, cancelMutation } = useOrders(statusFilter ? { status: statusFilter } : {})
  const orders = listQuery.data ?? []

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-5xl mx-auto p-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-6">Orders</h1>

        {/* Status tabs */}
        <div className="flex gap-2 flex-wrap mb-5">
          {STATUS_TABS.map((tab) => (
            <button
              key={tab.value}
              onClick={() => setStatusFilter(tab.value)}
              className={`text-xs px-3 py-1.5 rounded-full border transition-colors ${statusFilter === tab.value ? 'bg-amber-500 text-white border-amber-500' : 'bg-white text-gray-700 border-gray-200 hover:border-amber-300'}`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {listQuery.isLoading && <p className="text-sm text-gray-400">Loading…</p>}

        <div className="bg-white rounded-lg border shadow-sm divide-y">
          {orders.length === 0 && !listQuery.isLoading && (
            <p className="p-4 text-sm text-gray-400">No orders found.</p>
          )}
          {orders.map((order) => (
            <div key={order.id} className="p-4 flex items-start justify-between gap-4">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${STATUS_COLORS[order.status]}`}>
                    {order.status}
                  </span>
                  <span className="text-xs text-gray-400 uppercase">{order.source === 'customer_qr' ? 'QR' : 'POS'}</span>
                </div>
                <ul className="text-sm text-gray-700 space-y-0.5">
                  {order.items.map((item) => (
                    <li key={item.id} className="flex gap-2">
                      <span>{item.nameSnapshot}</span>
                      <span className="text-gray-400">×{item.quantity}</span>
                    </li>
                  ))}
                </ul>
                {order.note && <p className="text-xs text-gray-400 mt-1 italic">{order.note}</p>}
              </div>
              <div className="text-right shrink-0">
                <p className="text-sm font-bold text-amber-600">{formatIDR(order.totalMinor)}</p>
                <p className="text-xs text-gray-400 mt-0.5">
                  {new Date(order.createdAt).toLocaleString('id-ID', { dateStyle: 'short', timeStyle: 'short' })}
                </p>
                {order.status !== 'completed' && order.status !== 'cancelled' && (
                  <button
                    onClick={() => { if (confirm('Cancel this order?')) cancelMutation.mutate(order.id) }}
                    className="text-xs text-red-500 hover:text-red-700 mt-1"
                  >
                    Cancel
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
