import { createRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Route as authRoute } from './_auth'
import { useTables } from '../hooks/useTables'
import { useCategories } from '../hooks/useCategories'
import { useMenuItems } from '../hooks/useMenuItems'
import { useOrders } from '../hooks/useOrders'
import { apiFetch } from '../lib/api'
import type { MenuItem, Order, PaymentMethod } from '../types/api'
import { formatIDR } from '../lib/currency'

export const Route = createRoute({
  getParentRoute: () => authRoute,
  path: '/cashier',
  component: CashierPage,
})

interface CartItem {
  menuItem: MenuItem
  quantity: number
}

const paymentSchema = z.object({
  method: z.enum(['cash', 'qris', 'card']),
  amountMinor: z.coerce.number().int().min(1, 'Required'),
  receivedMinor: z.coerce.number().int().min(0),
})
type PaymentForm = z.infer<typeof paymentSchema>

function CashierPage() {
  const [selectedTableId, setSelectedTableId] = useState<string | null>(null)
  const [selectedCategoryId, setSelectedCategoryId] = useState<string | null>(null)
  const [cart, setCart] = useState<CartItem[]>([])
  const [placedOrder, setPlacedOrder] = useState<Order | null>(null)
  const [paymentError, setPaymentError] = useState<string | null>(null)
  const [paymentDone, setPaymentDone] = useState(false)

  const { listQuery: tablesQuery } = useTables()
  const { listQuery: catQuery } = useCategories()
  const { listQuery: itemsQuery } = useMenuItems(
    selectedCategoryId ? { categoryId: selectedCategoryId, availableOnly: true } : { availableOnly: true }
  )
  const { createMutation } = useOrders()

  const payForm = useForm<PaymentForm>({
    resolver: zodResolver(paymentSchema),
    defaultValues: { method: 'cash', amountMinor: 0, receivedMinor: 0 },
  })

  function addToCart(item: MenuItem) {
    setCart((prev) => {
      const existing = prev.find((c) => c.menuItem.id === item.id)
      if (existing) {
        return prev.map((c) => c.menuItem.id === item.id ? { ...c, quantity: c.quantity + 1 } : c)
      }
      return [...prev, { menuItem: item, quantity: 1 }]
    })
  }

  function removeFromCart(itemId: string) {
    setCart((prev) => prev.filter((c) => c.menuItem.id !== itemId))
  }

  function adjustQty(itemId: string, delta: number) {
    setCart((prev) =>
      prev
        .map((c) => c.menuItem.id === itemId ? { ...c, quantity: c.quantity + delta } : c)
        .filter((c) => c.quantity > 0)
    )
  }

  const cartTotal = cart.reduce((sum, c) => sum + c.menuItem.priceMinor * c.quantity, 0)

  async function placeOrder() {
    if (!selectedTableId || cart.length === 0) return
    const result = await createMutation.mutateAsync({
      tableId: selectedTableId,
      items: cart.map((c) => ({ menuItemId: c.menuItem.id, quantity: c.quantity })),
    })
    setPlacedOrder(result.order)
    setCart([])
    payForm.reset({ method: 'cash', amountMinor: result.order.totalMinor, receivedMinor: result.order.totalMinor })
  }

  async function recordPayment(data: PaymentForm) {
    if (!placedOrder) return
    setPaymentError(null)
    try {
      await apiFetch(`/api/orders/${placedOrder.id}/payments`, {
        method: 'POST',
        body: JSON.stringify(data),
      })
      setPaymentDone(true)
    } catch (e: unknown) {
      setPaymentError(e instanceof Error ? e.message : 'Payment failed')
    }
  }

  function resetAll() {
    setSelectedTableId(null)
    setPlacedOrder(null)
    setPaymentDone(false)
    setPaymentError(null)
    setCart([])
  }

  const tables = tablesQuery.data ?? []
  const categories = catQuery.data ?? []
  const items = itemsQuery.data ?? []

  // ── Step 3: payment done ───────────────────────────────────────────────────
  if (paymentDone) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-6">
        <div className="bg-white rounded-xl shadow-sm p-8 max-w-sm w-full text-center">
          <p className="text-4xl mb-4">✓</p>
          <h2 className="text-xl font-bold text-gray-800 mb-2">Payment recorded</h2>
          <p className="text-sm text-gray-500 mb-6">Order is complete.</p>
          <button onClick={resetAll} className="w-full bg-amber-500 text-white py-2 rounded-lg hover:bg-amber-600">
            New order
          </button>
        </div>
      </div>
    )
  }

  // ── Step 2: order placed, record payment ──────────────────────────────────
  if (placedOrder) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-6">
        <div className="bg-white rounded-xl shadow-sm p-6 max-w-sm w-full">
          <h2 className="font-bold text-gray-800 text-lg mb-1">Order placed</h2>
          <p className="text-sm text-gray-500 mb-4">Total: {formatIDR(placedOrder.totalMinor)}</p>

          <form onSubmit={payForm.handleSubmit(recordPayment)} className="space-y-3">
            <div>
              <label className="text-xs font-medium text-gray-600">Method</label>
              <select {...payForm.register('method')} className="w-full border rounded-md px-3 py-2 text-sm mt-1">
                <option value="cash">Cash</option>
                <option value="qris">QRIS</option>
                <option value="card">Card</option>
              </select>
            </div>
            <div>
              <label className="text-xs font-medium text-gray-600">Amount charged (IDR)</label>
              <input {...payForm.register('amountMinor')} type="number" className="w-full border rounded-md px-3 py-2 text-sm mt-1" />
            </div>
            <div>
              <label className="text-xs font-medium text-gray-600">Cash received (IDR)</label>
              <input {...payForm.register('receivedMinor')} type="number" className="w-full border rounded-md px-3 py-2 text-sm mt-1" />
            </div>
            {paymentError && <p className="text-red-500 text-xs">{paymentError}</p>}
            <div className="flex gap-2">
              <button type="submit" className="flex-1 bg-amber-500 text-white py-2 rounded-md text-sm hover:bg-amber-600">
                Record payment
              </button>
              <button type="button" onClick={resetAll} className="flex-1 bg-gray-100 text-gray-700 py-2 rounded-md text-sm hover:bg-gray-200">
                Cancel
              </button>
            </div>
          </form>
        </div>
      </div>
    )
  }

  // ── Step 1: pick table + menu ──────────────────────────────────────────────
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-5xl mx-auto p-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-6">Cashier POS</h1>

        <div className="flex gap-6">
          {/* Left: table + menu */}
          <div className="flex-1 space-y-4">
            {/* Table picker */}
            <div className="bg-white rounded-lg border shadow-sm p-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">Select table</label>
              <select
                value={selectedTableId ?? ''}
                onChange={(e) => setSelectedTableId(e.target.value || null)}
                className="w-full border rounded-md px-3 py-2 text-sm"
              >
                <option value="">— pick a table —</option>
                {tables.map((t) => (
                  <option key={t.id} value={t.id}>{t.label}</option>
                ))}
              </select>
            </div>

            {/* Category filter */}
            <div className="flex gap-2 flex-wrap">
              <button
                onClick={() => setSelectedCategoryId(null)}
                className={`text-xs px-3 py-1.5 rounded-full border ${!selectedCategoryId ? 'bg-amber-500 text-white border-amber-500' : 'bg-white text-gray-700 border-gray-200 hover:border-amber-300'}`}
              >
                All
              </button>
              {categories.map((cat) => (
                <button
                  key={cat.id}
                  onClick={() => setSelectedCategoryId(cat.id)}
                  className={`text-xs px-3 py-1.5 rounded-full border ${selectedCategoryId === cat.id ? 'bg-amber-500 text-white border-amber-500' : 'bg-white text-gray-700 border-gray-200 hover:border-amber-300'}`}
                >
                  {cat.name}
                </button>
              ))}
            </div>

            {/* Menu grid */}
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
              {items.map((item) => (
                <button
                  key={item.id}
                  onClick={() => addToCart(item)}
                  className="bg-white rounded-lg border shadow-sm p-3 text-left hover:border-amber-300 hover:shadow-md transition-all"
                >
                  <p className="text-sm font-semibold text-gray-800 truncate">{item.name}</p>
                  <p className="text-xs text-amber-600 font-bold mt-1">{formatIDR(item.priceMinor)}</p>
                </button>
              ))}
              {items.length === 0 && (
                <p className="col-span-3 text-sm text-gray-400 py-4 text-center">No items available</p>
              )}
            </div>
          </div>

          {/* Right: cart */}
          <div className="w-72 shrink-0">
            <div className="bg-white rounded-lg border shadow-sm p-4 sticky top-6">
              <h2 className="font-semibold text-gray-700 mb-3">Cart</h2>
              {cart.length === 0 && (
                <p className="text-sm text-gray-400">Add items from the menu.</p>
              )}
              <ul className="space-y-2 mb-4">
                {cart.map((c) => (
                  <li key={c.menuItem.id} className="flex items-center gap-2">
                    <div className="flex-1 min-w-0">
                      <p className="text-sm truncate text-gray-800">{c.menuItem.name}</p>
                      <p className="text-xs text-amber-600">{formatIDR(c.menuItem.priceMinor)}</p>
                    </div>
                    <div className="flex items-center gap-1">
                      <button onClick={() => adjustQty(c.menuItem.id, -1)} className="w-5 h-5 text-xs border rounded text-gray-600 hover:bg-gray-50">−</button>
                      <span className="text-xs w-4 text-center">{c.quantity}</span>
                      <button onClick={() => adjustQty(c.menuItem.id, 1)} className="w-5 h-5 text-xs border rounded text-gray-600 hover:bg-gray-50">+</button>
                      <button onClick={() => removeFromCart(c.menuItem.id)} className="text-red-400 hover:text-red-600 text-xs ml-1">✕</button>
                    </div>
                  </li>
                ))}
              </ul>
              {cart.length > 0 && (
                <>
                  <div className="border-t pt-3 mb-3 flex justify-between">
                    <span className="text-sm font-semibold text-gray-700">Total</span>
                    <span className="text-sm font-bold text-amber-600">{formatIDR(cartTotal)}</span>
                  </div>
                  <button
                    onClick={placeOrder}
                    disabled={!selectedTableId || createMutation.isPending}
                    className="w-full bg-amber-500 text-white py-2 rounded-md text-sm font-medium hover:bg-amber-600 disabled:opacity-60"
                  >
                    {!selectedTableId ? 'Select a table first' : 'Place order'}
                  </button>
                </>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
