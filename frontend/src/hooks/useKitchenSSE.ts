import { useEffect, useReducer } from 'react'
import type { Order, KitchenSSEEvent } from '../types/api'

type State = { orders: Order[] }
type Action =
  | { type: 'order.created'; order: Order }
  | { type: 'order.status_changed'; order: Order }
  | { type: 'order.cancelled'; order: Order }
  | { type: 'init'; orders: Order[] }

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'init':
      return { orders: action.orders }
    case 'order.created':
      return { orders: [action.order, ...state.orders] }
    case 'order.status_changed':
    case 'order.cancelled':
      return {
        orders: state.orders.map((o) => (o.id === action.order.id ? action.order : o)),
      }
  }
}

export function useKitchenSSE(accessToken: string | null, initialOrders: Order[] = []) {
  const [state, dispatch] = useReducer(reducer, { orders: initialOrders })

  useEffect(() => {
    dispatch({ type: 'init', orders: initialOrders })
  }, [initialOrders]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!accessToken) return

    const es = new EventSource(`/api/sse/kitchen?token=${encodeURIComponent(accessToken)}`)

    const handleEvent = (e: MessageEvent) => {
      try {
        const evt = JSON.parse(e.data) as KitchenSSEEvent
        dispatch({ type: evt.type, order: evt.payload })
      } catch {
        // ignore malformed events
      }
    }

    es.addEventListener('order.created', handleEvent as EventListener)
    es.addEventListener('order.status_changed', handleEvent as EventListener)
    es.addEventListener('order.cancelled', handleEvent as EventListener)
    es.onmessage = handleEvent

    return () => es.close()
  }, [accessToken])

  return state.orders
}
