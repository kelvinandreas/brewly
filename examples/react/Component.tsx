// Canonical React component pattern.
// Rules: one component per file; named + default export of the same identifier.
// Server data through hooks only — no fetch in components.
import { useThings, useCreateThing } from '@/hooks/useThings'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const createThingSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100),
})

type CreateThingForm = z.infer<typeof createThingSchema>

export function ThingList() {
  const { data: things, isLoading, isError } = useThings()
  const createThing = useCreateThing()

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateThingForm>({ resolver: zodResolver(createThingSchema) })

  const onSubmit = (values: CreateThingForm) => {
    createThing.mutate(values, { onSuccess: () => reset() })
  }

  if (isLoading) return <p className="p-4 text-gray-500">Loading…</p>
  if (isError) return <p className="p-4 text-red-500">Failed to load things.</p>

  return (
    <div className="space-y-4 p-4">
      <form onSubmit={handleSubmit(onSubmit)} className="flex gap-2">
        <input
          {...register('name')}
          placeholder="Thing name"
          className="flex-1 rounded border px-3 py-2 text-sm"
        />
        {errors.name && <p className="text-xs text-red-500">{errors.name.message}</p>}
        <button
          type="submit"
          disabled={createThing.isPending}
          className="rounded bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
        >
          Add
        </button>
      </form>

      <ul className="divide-y rounded border">
        {things?.map((thing) => (
          <li key={thing.id} className="flex items-center justify-between px-4 py-3">
            <span className="text-sm font-medium">{thing.name}</span>
            <span className="text-xs text-gray-400">{thing.status}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}

export default ThingList
