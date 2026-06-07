import { createRoute } from '@tanstack/react-router'
import { Route as authRoute } from './_auth'
import { useAuth } from '../hooks/useAuth'

export const Route = createRoute({
  getParentRoute: () => authRoute,
  path: '/staff',
  component: StaffPage,
})

function StaffPage() {
  const { user } = useAuth()

  if (user?.role !== 'owner') {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-gray-500">Owner access required.</p>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-3xl mx-auto">
        <h1 className="text-2xl font-bold text-gray-900 mb-6">Staff</h1>
        <p className="text-sm text-gray-400">Staff management coming soon — M2 adds the full CRUD UI.</p>
      </div>
    </div>
  )
}
