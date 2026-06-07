import { createRoute } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Route as rootRoute } from './__root'
import { setTableToken, getTableToken } from '../lib/tableAuth'
import { useCustomerOrder } from '../hooks/useCustomerOrder'
import { useYouTubeSearch, useSubmitSongRequest } from '../hooks/useCustomerSong'
import type { CustomerMenuResponse, CustomerMenuItem, YouTubeVideoResult } from '../types/api'
import { formatIDR } from '../lib/currency'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/table/$tableId',
  component: CustomerMenuPage,
})

interface CartItem {
  item: CustomerMenuItem
  quantity: number
}

async function fetchCustomerMenu(token: string): Promise<CustomerMenuResponse> {
  const res = await fetch('/api/customer/menu', {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) throw new Error('Failed to load menu')
  const body = await res.json()
  return body.data as CustomerMenuResponse
}

type Tab = 'menu' | 'songs'

function CustomerMenuPage() {
  const { tableId } = Route.useParams()
  const [ready, setReady] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<Tab>('menu')
  const [cart, setCart] = useState<CartItem[]>([])
  const [cartOpen, setCartOpen] = useState(false)
  const [orderSuccess, setOrderSuccess] = useState(false)

  const { placeOrderMutation, myOrdersQuery } = useCustomerOrder(tableId)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const token = params.get('token')
    if (token) {
      setTableToken(token)
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

  function addToCart(item: CustomerMenuItem) {
    setCart((prev) => {
      const existing = prev.find((c) => c.item.id === item.id)
      if (existing)
        return prev.map((c) => (c.item.id === item.id ? { ...c, quantity: c.quantity + 1 } : c))
      return [...prev, { item, quantity: 1 }]
    })
  }

  function adjustQty(itemId: string, delta: number) {
    setCart((prev) =>
      prev
        .map((c) => (c.item.id === itemId ? { ...c, quantity: c.quantity + delta } : c))
        .filter((c) => c.quantity > 0),
    )
  }

  const cartTotal = cart.reduce((sum, c) => sum + c.item.priceMinor * c.quantity, 0)
  const cartCount = cart.reduce((sum, c) => sum + c.quantity, 0)

  async function submitOrder() {
    await placeOrderMutation.mutateAsync({
      items: cart.map((c) => ({ menuItemId: c.item.id, quantity: c.quantity })),
    })
    setCart([])
    setCartOpen(false)
    setOrderSuccess(true)
    setTimeout(() => setOrderSuccess(false), 4000)
  }

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
          <p className="text-sm text-gray-500 mt-1">
            Your QR code may have expired. Please scan again.
          </p>
        </div>
      </div>
    )
  }

  const categories = menuQuery.data?.categories ?? []
  const myOrders = myOrdersQuery.data ?? []

  return (
    <div className="min-h-screen bg-amber-50 pb-24">
      <header className="bg-amber-600 text-white px-4 py-4 sticky top-0 z-10 shadow">
        <div className="flex items-center justify-between max-w-lg mx-auto">
          <div>
            <h1 className="text-lg font-bold">Brewly</h1>
            <p className="text-xs text-amber-200">Table {tableId}</p>
          </div>
          {activeTab === 'menu' && cartCount > 0 && (
            <button
              onClick={() => setCartOpen(true)}
              className="relative bg-white text-amber-600 px-4 py-1.5 rounded-full text-sm font-semibold shadow"
            >
              Cart ({cartCount}) · {formatIDR(cartTotal)}
            </button>
          )}
        </div>

        {/* Tab switcher */}
        <div className="flex gap-1 mt-3 bg-amber-700/50 p-1 rounded-lg max-w-lg mx-auto">
          {(['menu', 'songs'] as Tab[]).map((t) => (
            <button
              key={t}
              onClick={() => setActiveTab(t)}
              className={`flex-1 py-1.5 text-sm font-medium rounded-md transition-colors ${
                activeTab === t ? 'bg-white text-amber-700' : 'text-amber-100 hover:text-white'
              }`}
            >
              {t === 'menu' ? 'Menu' : 'Request Song'}
            </button>
          ))}
        </div>
      </header>

      {orderSuccess && (
        <div className="max-w-lg mx-auto px-4 pt-4">
          <div className="bg-green-100 text-green-700 text-sm rounded-lg px-4 py-3 flex items-center gap-2">
            <span>✓</span> Order placed! Kitchen is on it.
          </div>
        </div>
      )}

      {activeTab === 'menu' ? (
        <div className="max-w-lg mx-auto px-4 py-6 space-y-8">
          {categories.length === 0 && (
            <p className="text-center text-sm text-gray-400">No items available right now.</p>
          )}
          {categories.map((cat) => (
            <section key={cat.id}>
              <h2 className="text-base font-bold text-gray-800 mb-3 border-b border-amber-200 pb-1">
                {cat.name}
              </h2>
              {cat.items.length === 0 && (
                <p className="text-sm text-gray-400 pl-1">No available items in this category.</p>
              )}
              <div className="grid grid-cols-2 gap-3">
                {cat.items.map((item) => {
                  const inCart = cart.find((c) => c.item.id === item.id)
                  return (
                    <div key={item.id} className="bg-white rounded-xl shadow-sm overflow-hidden">
                      {item.imageUrl ? (
                        <img
                          src={item.imageUrl}
                          alt={item.name}
                          className="w-full h-28 object-cover"
                        />
                      ) : (
                        <div className="w-full h-28 bg-amber-100 flex items-center justify-center">
                          <span className="text-3xl">☕</span>
                        </div>
                      )}
                      <div className="p-2">
                        <p className="text-sm font-semibold text-gray-800 leading-tight">
                          {item.name}
                        </p>
                        {item.description && (
                          <p className="text-xs text-gray-400 mt-0.5 line-clamp-2">
                            {item.description}
                          </p>
                        )}
                        <div className="flex items-center justify-between mt-1">
                          <p className="text-xs font-bold text-amber-600">
                            {formatIDR(item.priceMinor)}
                          </p>
                          {inCart ? (
                            <div className="flex items-center gap-1">
                              <button
                                onClick={() => adjustQty(item.id, -1)}
                                className="w-5 h-5 text-xs border rounded text-gray-600"
                              >
                                −
                              </button>
                              <span className="text-xs w-4 text-center">{inCart.quantity}</span>
                              <button
                                onClick={() => adjustQty(item.id, 1)}
                                className="w-5 h-5 text-xs border rounded text-gray-600"
                              >
                                +
                              </button>
                            </div>
                          ) : (
                            <button
                              onClick={() => addToCart(item)}
                              className="text-xs bg-amber-500 text-white px-2 py-0.5 rounded-full hover:bg-amber-600"
                            >
                              Add
                            </button>
                          )}
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            </section>
          ))}

          {/* My orders */}
          {myOrders.length > 0 && (
            <section>
              <h2 className="text-base font-bold text-gray-800 mb-3 border-b border-amber-200 pb-1">
                My Orders
              </h2>
              <div className="space-y-2">
                {myOrders.map((order) => (
                  <div
                    key={order.id}
                    className="bg-white rounded-lg px-3 py-2 flex items-center justify-between shadow-sm"
                  >
                    <div>
                      <p className="text-xs text-gray-500">
                        {order.items.map((i) => `${i.nameSnapshot} ×${i.quantity}`).join(', ')}
                      </p>
                      <p className="text-xs font-bold text-amber-600 mt-0.5">
                        {formatIDR(order.totalMinor)}
                      </p>
                    </div>
                    <span
                      className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                        order.status === 'ready'
                          ? 'bg-green-100 text-green-700'
                          : order.status === 'completed'
                            ? 'bg-gray-100 text-gray-600'
                            : order.status === 'cancelled'
                              ? 'bg-red-100 text-red-600'
                              : 'bg-yellow-100 text-yellow-700'
                      }`}
                    >
                      {order.status}
                    </span>
                  </div>
                ))}
              </div>
            </section>
          )}
        </div>
      ) : (
        <SongsTab tableId={tableId} />
      )}

      {/* Cart drawer */}
      {cartOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-50 flex items-end"
          onClick={() => setCartOpen(false)}
        >
          <div
            className="bg-white w-full rounded-t-2xl p-5 max-h-[80vh] overflow-y-auto"
            onClick={(e) => e.stopPropagation()}
          >
            <h3 className="font-bold text-gray-800 text-lg mb-4">Your order</h3>
            <ul className="space-y-3 mb-4">
              {cart.map((c) => (
                <li key={c.item.id} className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-800">{c.item.name}</p>
                    <p className="text-xs text-amber-600">{formatIDR(c.item.priceMinor)}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => adjustQty(c.item.id, -1)}
                      className="w-6 h-6 border rounded text-sm"
                    >
                      −
                    </button>
                    <span className="text-sm w-5 text-center">{c.quantity}</span>
                    <button
                      onClick={() => adjustQty(c.item.id, 1)}
                      className="w-6 h-6 border rounded text-sm"
                    >
                      +
                    </button>
                  </div>
                </li>
              ))}
            </ul>
            <div className="border-t pt-3 mb-4 flex justify-between">
              <span className="font-semibold text-gray-700">Total</span>
              <span className="font-bold text-amber-600">{formatIDR(cartTotal)}</span>
            </div>
            <button
              onClick={submitOrder}
              disabled={placeOrderMutation.isPending}
              className="w-full bg-amber-500 text-white py-3 rounded-xl text-base font-semibold hover:bg-amber-600 disabled:opacity-60"
            >
              {placeOrderMutation.isPending ? 'Placing order…' : 'Place order'}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

function SongsTab({ tableId }: { tableId: string }) {
  const [query, setQuery] = useState('')
  const [debouncedQ, setDebouncedQ] = useState('')
  const [note, setNote] = useState('')
  const [submitSuccess, setSubmitSuccess] = useState<string | null>(null)
  const [submitError, setSubmitError] = useState<string | null>(null)

  const submitMutation = useSubmitSongRequest(tableId)

  useEffect(() => {
    const t = setTimeout(() => setDebouncedQ(query), 500)
    return () => clearTimeout(t)
  }, [query])

  const searchQuery = useYouTubeSearch(debouncedQ)

  async function handleRequest(video: YouTubeVideoResult) {
    setSubmitError(null)
    try {
      await submitMutation.mutateAsync({
        videoId: video.videoId,
        title: video.title,
        channelName: video.channelName,
        thumbnailUrl: video.thumbnailUrl,
        note,
      })
      setNote('')
      setSubmitSuccess(`"${video.title}" added to the queue!`)
      setTimeout(() => setSubmitSuccess(null), 4000)
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Failed to submit'
      setSubmitError(msg.includes('rate_limited') ? 'You have too many queued requests.' : msg)
    }
  }

  return (
    <div className="max-w-lg mx-auto px-4 py-6 space-y-4">
      {submitSuccess && (
        <div className="bg-green-100 text-green-700 text-sm rounded-lg px-4 py-3">
          ✓ {submitSuccess}
        </div>
      )}
      {submitError && (
        <div className="bg-red-100 text-red-600 text-sm rounded-lg px-4 py-3">{submitError}</div>
      )}

      <div>
        <input
          type="text"
          placeholder="Search for a song…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="w-full border rounded-xl px-4 py-3 text-sm shadow-sm focus:outline-none focus:ring-2 focus:ring-amber-400"
        />
      </div>

      <div>
        <input
          type="text"
          placeholder="Optional note for the DJ"
          value={note}
          onChange={(e) => setNote(e.target.value)}
          className="w-full border rounded-xl px-4 py-2 text-sm shadow-sm focus:outline-none focus:ring-2 focus:ring-amber-400"
        />
      </div>

      {searchQuery.isLoading && <p className="text-center text-sm text-gray-400">Searching…</p>}

      {searchQuery.data && searchQuery.data.length === 0 && debouncedQ.length >= 2 && (
        <p className="text-center text-sm text-gray-400">No results for "{debouncedQ}"</p>
      )}

      <div className="space-y-3">
        {searchQuery.data?.map((video) => (
          <div
            key={video.videoId}
            className="bg-white rounded-xl shadow-sm flex items-center gap-3 p-3"
          >
            {video.thumbnailUrl && (
              <img
                src={video.thumbnailUrl}
                alt={video.title}
                className="w-16 h-12 object-cover rounded-lg flex-shrink-0"
              />
            )}
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-gray-800 line-clamp-2 leading-tight">
                {video.title}
              </p>
              <p className="text-xs text-gray-400 mt-0.5">{video.channelName}</p>
            </div>
            <button
              onClick={() => handleRequest(video)}
              disabled={submitMutation.isPending}
              className="flex-shrink-0 text-xs bg-amber-500 text-white px-3 py-1.5 rounded-full font-medium hover:bg-amber-600 disabled:opacity-60"
            >
              Request
            </button>
          </div>
        ))}
      </div>

      {!query && (
        <p className="text-center text-xs text-gray-400 pt-4">
          Type a song name or artist to search YouTube
        </p>
      )}
    </div>
  )
}
