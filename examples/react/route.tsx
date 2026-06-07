// Canonical TanStack Router file-based route pattern.
// Filename convention: src/routes/_auth/things/index.tsx → path /_auth/things/
// Protected routes nest under _auth.tsx whose beforeLoad redirects to /login.
import { createFileRoute } from '@tanstack/react-router'
import { thingKeys } from '@/hooks/useThings'
import { api } from '@/lib/api'
import { ThingList } from '@/components/ThingList'

export const Route = createFileRoute('/_auth/things/')({
  // Loader runs before the component renders. Prefill the cache from the server.
  loader: ({ context: { queryClient } }) =>
    queryClient.ensureQueryData({
      queryKey: thingKeys.list(),
      queryFn: () => api.get('/api/things'),
    }),

  component: ThingsPage,
})

function ThingsPage() {
  return (
    <div className="mx-auto max-w-2xl py-8">
      <h1 className="mb-6 text-2xl font-bold">Things</h1>
      <ThingList />
    </div>
  )
}
