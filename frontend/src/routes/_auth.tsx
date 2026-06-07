import { createRoute, Outlet, redirect } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { refreshAccess, getAccessToken } from '../lib/auth'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  id: '_auth',
  beforeLoad: async () => {
    const token = getAccessToken() ?? (await refreshAccess())
    if (!token) throw redirect({ to: '/login' as string })
  },
  component: () => <Outlet />,
})
