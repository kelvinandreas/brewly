import { createRoute, redirect } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { getAccessToken, refreshAccess } from '../lib/auth'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  beforeLoad: async () => {
    const token = getAccessToken() ?? (await refreshAccess())
    if (token) throw redirect({ to: '/dashboard' as string })
    throw redirect({ to: '/login' as string })
  },
})
