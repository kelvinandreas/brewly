import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createRouter, RouterProvider } from '@tanstack/react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Route as rootRoute } from './routes/__root'
import { Route as indexRoute } from './routes/index'
import { Route as loginRoute } from './routes/login'
import { Route as authRoute } from './routes/_auth'
import { Route as dashboardRoute } from './routes/_auth.dashboard'
import { Route as staffRoute } from './routes/_auth.staff'
import { Route as menuRoute } from './routes/_auth.menu'
import { Route as tablesRoute } from './routes/_auth.tables'
import { Route as kitchenRoute } from './routes/_auth.kitchen'
import { Route as cashierRoute } from './routes/_auth.cashier'
import { Route as ordersRoute } from './routes/_auth.orders'
import { Route as songQueueRoute } from './routes/_auth.song-queue'
import { Route as reportsRoute } from './routes/_auth.reports'
import { Route as tableCustomerRoute } from './routes/table.$tableId'
import './index.css'

const authTree = authRoute.addChildren([dashboardRoute, staffRoute, menuRoute, tablesRoute, kitchenRoute, cashierRoute, ordersRoute, songQueueRoute, reportsRoute])

const routeTree = rootRoute.addChildren([indexRoute, loginRoute, authTree, tableCustomerRoute])

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 1000 * 60 },
  },
})

const router = createRouter({
  routeTree,
  context: { queryClient },
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

const rootElement = document.getElementById('root')!

createRoot(rootElement).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
)
