import { createRoute } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Route as rootRoute } from './__root'
import { setTableToken, getTableToken } from '../lib/tableAuth'
import type { CustomerMenuResponse } from '../types/api'
import { formatIDR } from '../lib/currency'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/table/$tableId',
  component: CustomerMenuPage,
})

async function fetchCustomerMenu(token: string): Promise<CustomerMenuResponse> {
  const res = await fetch('/api/customer/menu', {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) throw new Error('Failed to load menu')
  const body = await res.json()
  return body.data as CustomerMenuResponse
}

function CustomerMenuPage() {
  const { tableId } = Route.useParams()
  const [ready, setReady] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const token = params.get('token')
    if (token) {
      setTableToken(token)
      // Strip token from URL without a history entry
      window.history.replaceState(null, '', `/table/${tableId}`)
    }
    if (getTableToken()) {
      setReady(true)
    } else {
      setError('Invalid or missing table token. Please scan the QR code again.')
    }
  }, [tableId])

  const menuQuery = useQuery({
    queryKey: ['customer-menu', tableId],
    queryFn: () => fetchCustomerMenu(getTableToken()!),
    enabled: ready,
    retry: false,
  })

  if (error) {
    return (
      <div className="min-h-screen bg-amber-50 flex items-center justify-center p-6">
        <div className="text-center">
          <p className="text-red-600 font-medium mb-2">Access denied</p>
          <p className="text-sm text-gray-500">{error}</p>
        </div>
      </div>
    )
  }

  if (!ready || menuQuery.isLoading) {
    return (
      <div className="min-h-screen bg-amber-50 flex items-center justify-center">
        <p className="text-amber-600 text-sm">Loading menu…</p>
      </div>
    )
  }

  if (menuQuery.isError) {
    return (
      <div className="min-h-screen bg-amber-50 flex items-center justify-center p-6">
        <div className="text-center">
          <p className="text-red-600 font-medium">Could not load menu</p>
          <p className="text-sm text-gray-500 mt-1">Your QR code may have expired. Please scan again.</p>
        </div>
      </div>
    )
  }

  const categories = menuQuery.data?.categories ?? []

  return (
    <div className="min-h-screen bg-amber-50">
      <header className="bg-amber-600 text-white px-4 py-4 sticky top-0 z-10 shadow">
        <h1 className="text-lg font-bold">Brewly</h1>
        <p className="text-xs text-amber-200">Table {tableId}</p>
      </header>

      <div className="max-w-lg mx-auto px-4 py-6 space-y-8">
        {categories.length === 0 && (
          <p className="text-center text-sm text-gray-400">No items available right now.</p>
        )}
        {categories.map((cat) => (
          <section key={cat.id}>
            <h2 className="text-base font-bold text-gray-800 mb-3 border-b border-amber-200 pb-1">{cat.name}</h2>
            {cat.items.length === 0 && (
              <p className="text-sm text-gray-400 pl-1">No available items in this category.</p>
            )}
            <div className="grid grid-cols-2 gap-3">
              {cat.items.map((item) => (
                <div key={item.id} className="bg-white rounded-xl shadow-sm overflow-hidden">
                  {item.imageUrl ? (
                    <img src={item.imageUrl} alt={item.name} className="w-full h-28 object-cover" />
                  ) : (
                    <div className="w-full h-28 bg-amber-100 flex items-center justify-center">
                      <span className="text-3xl">☕</span>
                    </div>
                  )}
                  <div className="p-2">
                    <p className="text-sm font-semibold text-gray-800 leading-tight">{item.name}</p>
                    {item.description && (
                      <p className="text-xs text-gray-400 mt-0.5 line-clamp-2">{item.description}</p>
                    )}
                    <p className="text-xs font-bold text-amber-600 mt-1">{formatIDR(item.priceMinor)}</p>
                  </div>
                </div>
              ))}
            </div>
          </section>
        ))}
      </div>
    </div>
  )
}
