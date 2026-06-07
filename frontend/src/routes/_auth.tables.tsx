import { createRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Route as authRoute } from './_auth'
import { useTables } from '../hooks/useTables'
import type { Table, TableCreateResponse, TableRegenerateResponse } from '../types/api'

export const Route = createRoute({
  getParentRoute: () => authRoute,
  path: '/tables',
  component: TablesPage,
})

const tableSchema = z.object({
  label: z.string().min(1, 'Required'),
})
type TableForm = z.infer<typeof tableSchema>

interface QRModalData {
  tableLabel: string
  qrToken: string
  qrUrl: string
  qrImageUrl: string
}

function TablesPage() {
  const { listQuery, createMutation, updateMutation, deleteMutation, regenerateTokenMutation } = useTables()
  const [qrModal, setQRModal] = useState<QRModalData | null>(null)
  const [editingTable, setEditingTable] = useState<Table | null>(null)

  const form = useForm<TableForm>({ resolver: zodResolver(tableSchema), defaultValues: { label: '' } })

  function openQR(data: { tableLabel: string; qrToken: string; qrUrl: string; tableId: string }) {
    setQRModal({
      tableLabel: data.tableLabel,
      qrToken: data.qrToken,
      qrUrl: data.qrUrl,
      qrImageUrl: `/api/tables/${data.tableId}/qr.png`,
    })
  }

  async function submitTable(data: TableForm) {
    if (editingTable) {
      await updateMutation.mutateAsync({ id: editingTable.id, label: data.label })
      setEditingTable(null)
    } else {
      const result: TableCreateResponse = await createMutation.mutateAsync(data)
      openQR({ tableLabel: result.table.label, qrToken: result.qrToken, qrUrl: result.qrUrl, tableId: result.table.id })
    }
    form.reset({ label: '' })
  }

  async function regenerate(table: Table) {
    const result: TableRegenerateResponse = await regenerateTokenMutation.mutateAsync(table.id)
    openQR({ tableLabel: table.label, qrToken: result.qrToken, qrUrl: result.qrUrl, tableId: table.id })
  }

  const tables = listQuery.data ?? []

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto p-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-6">Tables</h1>

        {/* Create / Edit form */}
        <div className="bg-white rounded-lg border shadow-sm p-4 mb-6">
          <h2 className="font-semibold text-gray-700 mb-3">{editingTable ? `Edit "${editingTable.label}"` : 'Add table'}</h2>
          <form onSubmit={form.handleSubmit(submitTable)} className="flex gap-3">
            <div className="flex-1">
              <input
                {...form.register('label')}
                placeholder="Table label (e.g. A1)"
                className="w-full border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-amber-400"
              />
              {form.formState.errors.label && (
                <p className="text-red-500 text-xs mt-1">{form.formState.errors.label.message}</p>
              )}
            </div>
            <button
              type="submit"
              disabled={createMutation.isPending || updateMutation.isPending}
              className="bg-amber-500 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-amber-600 disabled:opacity-60"
            >
              {editingTable ? 'Update' : 'Create'}
            </button>
            {editingTable && (
              <button
                type="button"
                onClick={() => { setEditingTable(null); form.reset({ label: '' }) }}
                className="bg-gray-100 text-gray-700 px-4 py-2 rounded-md text-sm hover:bg-gray-200"
              >
                Cancel
              </button>
            )}
          </form>
        </div>

        {/* Table list */}
        {listQuery.isLoading && <p className="text-sm text-gray-400">Loading…</p>}

        <div className="bg-white rounded-lg border shadow-sm divide-y">
          {tables.length === 0 && !listQuery.isLoading && (
            <p className="p-4 text-sm text-gray-400">No tables yet. Create one above.</p>
          )}
          {tables.map((table) => (
            <div key={table.id} className="flex items-center justify-between px-4 py-3">
              <div>
                <p className="font-medium text-gray-800">{table.label}</p>
                <p className="text-xs text-gray-400">v{table.tokenVersion}</p>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => { setEditingTable(table); form.reset({ label: table.label }) }}
                  className="text-xs border border-gray-300 text-gray-600 px-3 py-1.5 rounded hover:bg-gray-50"
                >
                  Edit
                </button>
                <button
                  onClick={() => regenerate(table)}
                  disabled={regenerateTokenMutation.isPending}
                  className="text-xs border border-amber-300 text-amber-700 px-3 py-1.5 rounded hover:bg-amber-50 disabled:opacity-60"
                >
                  Regen QR
                </button>
                <button
                  onClick={() => openQR({ tableLabel: table.label, qrToken: '', qrUrl: '', tableId: table.id })}
                  className="text-xs border border-blue-300 text-blue-700 px-3 py-1.5 rounded hover:bg-blue-50"
                >
                  View QR
                </button>
                <button
                  onClick={() => { if (confirm(`Delete table "${table.label}"?`)) deleteMutation.mutate(table.id) }}
                  className="text-xs border border-red-200 text-red-600 px-3 py-1.5 rounded hover:bg-red-50"
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>

        {/* QR Modal */}
        {qrModal && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setQRModal(null)}>
            <div className="bg-white rounded-xl p-6 max-w-sm w-full mx-4 shadow-xl" onClick={(e) => e.stopPropagation()}>
              <h3 className="font-bold text-gray-900 text-lg mb-1">{qrModal.tableLabel}</h3>
              <p className="text-xs text-gray-400 mb-4">Scan to open the customer menu</p>
              <div className="flex justify-center mb-4">
                <img
                  src={qrModal.qrImageUrl}
                  alt={`QR for ${qrModal.tableLabel}`}
                  className="w-48 h-48 border rounded"
                />
              </div>
              {qrModal.qrUrl && (
                <p className="text-xs text-gray-500 break-all mb-2">
                  <span className="font-medium">URL: </span>{qrModal.qrUrl}
                </p>
              )}
              {qrModal.qrToken && (
                <p className="text-xs text-gray-400 break-all">
                  <span className="font-medium">Token: </span>{qrModal.qrToken.slice(0, 40)}…
                </p>
              )}
              <button
                onClick={() => setQRModal(null)}
                className="mt-4 w-full bg-gray-100 text-gray-700 py-2 rounded-md text-sm hover:bg-gray-200"
              >
                Close
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
