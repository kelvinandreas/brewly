import { createRoute, Link } from '@tanstack/react-router'
import { Route as authRoute } from './_auth'
import { useAuth } from '../hooks/useAuth'

export const Route = createRoute({
  getParentRoute: () => authRoute,
  path: '/dashboard',
  component: DashboardPage,
})

function DashboardPage() {
  const { user, logout } = useAuth()

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-4xl mx-auto">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
            <p className="text-sm text-gray-500">
              Welcome, {user?.name ?? '…'} ({user?.role})
            </p>
          </div>
          <button
            onClick={() => logout.mutate()}
            disabled={logout.isPending}
            className="rounded-md bg-gray-200 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-300 disabled:opacity-60"
          >
            Sign out
          </button>
        </div>

        <div className="grid gap-4 grid-cols-1 sm:grid-cols-3">
          <Link
            to={'/menu' as string}
            className="rounded-lg border bg-white p-6 shadow-sm hover:shadow-md hover:border-amber-300 transition-shadow"
          >
            <h2 className="font-semibold text-gray-700">Menu</h2>
            <p className="mt-1 text-sm text-amber-500">Manage categories &amp; items</p>
          </Link>
          <Link
            to={'/tables' as string}
            className="rounded-lg border bg-white p-6 shadow-sm hover:shadow-md hover:border-amber-300 transition-shadow"
          >
            <h2 className="font-semibold text-gray-700">Tables</h2>
            <p className="mt-1 text-sm text-amber-500">Manage tables &amp; QR codes</p>
          </Link>
          <Link
            to={'/cashier' as string}
            className="rounded-lg border bg-white p-6 shadow-sm hover:shadow-md hover:border-amber-300 transition-shadow"
          >
            <h2 className="font-semibold text-gray-700">Cashier POS</h2>
            <p className="mt-1 text-sm text-amber-500">Place orders &amp; record payments</p>
          </Link>
          <Link
            to={'/kitchen' as string}
            className="rounded-lg border bg-white p-6 shadow-sm hover:shadow-md hover:border-amber-300 transition-shadow"
          >
            <h2 className="font-semibold text-gray-700">Kitchen</h2>
            <p className="mt-1 text-sm text-amber-500">Live order board (KDS)</p>
          </Link>
          <Link
            to={'/orders' as string}
            className="rounded-lg border bg-white p-6 shadow-sm hover:shadow-md hover:border-amber-300 transition-shadow"
          >
            <h2 className="font-semibold text-gray-700">Orders</h2>
            <p className="mt-1 text-sm text-amber-500">Order history &amp; management</p>
          </Link>
          <div className="rounded-lg border bg-white p-6 shadow-sm">
            <h2 className="font-semibold text-gray-700">Song Queue</h2>
            <p className="mt-1 text-sm text-gray-400">Coming in M4</p>
          </div>
        </div>
      </div>
    </div>
  )
}
