import { createRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Route as authRoute } from './_auth'
import { useCategories } from '../hooks/useCategories'
import { useMenuItems } from '../hooks/useMenuItems'
import type { Category, MenuItem } from '../types/api'
import { formatIDR } from '../lib/currency'

export const Route = createRoute({
  getParentRoute: () => authRoute,
  path: '/menu',
  component: MenuPage,
})

const categorySchema = z.object({
  name: z.string().min(1, 'Required'),
  displayOrder: z.coerce.number().int().min(0, 'Must be ≥ 0'),
})
type CategoryForm = z.infer<typeof categorySchema>

const menuItemSchema = z.object({
  name: z.string().min(1, 'Required'),
  description: z.string().optional(),
  priceMinor: z.coerce.number().int().min(0, 'Must be ≥ 0'),
  imageUrl: z.string().url('Must be a valid URL').optional().or(z.literal('')),
  isAvailable: z.boolean(),
})
type MenuItemForm = z.infer<typeof menuItemSchema>

function MenuPage() {
  const [selectedCategoryId, setSelectedCategoryId] = useState<string | null>(null)
  const [editingCategory, setEditingCategory] = useState<Category | null>(null)
  const [editingItem, setEditingItem] = useState<MenuItem | null>(null)
  const [showCategoryForm, setShowCategoryForm] = useState(false)
  const [showItemForm, setShowItemForm] = useState(false)

  const { listQuery: catQuery, createMutation: catCreate, updateMutation: catUpdate, deleteMutation: catDelete } = useCategories()
  const { listQuery: itemQuery, createMutation: itemCreate, updateMutation: itemUpdate, deleteMutation: itemDelete } = useMenuItems(
    selectedCategoryId ? { categoryId: selectedCategoryId } : {}
  )

  const catForm = useForm<CategoryForm>({ resolver: zodResolver(categorySchema), defaultValues: { name: '', displayOrder: 0 } })
  const itemForm = useForm<MenuItemForm>({ resolver: zodResolver(menuItemSchema), defaultValues: { name: '', description: '', priceMinor: 0, imageUrl: '', isAvailable: true } })

  function openCategoryEdit(cat: Category) {
    setEditingCategory(cat)
    catForm.reset({ name: cat.name, displayOrder: cat.displayOrder })
    setShowCategoryForm(true)
  }

  function openItemEdit(item: MenuItem) {
    setEditingItem(item)
    itemForm.reset({
      name: item.name,
      description: item.description ?? '',
      priceMinor: item.priceMinor,
      imageUrl: item.imageUrl ?? '',
      isAvailable: item.isAvailable,
    })
    setShowItemForm(true)
  }

  function closeCategoryForm() {
    setEditingCategory(null)
    setShowCategoryForm(false)
    catForm.reset({ name: '', displayOrder: 0 })
  }

  function closeItemForm() {
    setEditingItem(null)
    setShowItemForm(false)
    itemForm.reset({ name: '', description: '', priceMinor: 0, imageUrl: '', isAvailable: true })
  }

  async function submitCategory(data: CategoryForm) {
    if (editingCategory) {
      await catUpdate.mutateAsync({ id: editingCategory.id, ...data })
    } else {
      await catCreate.mutateAsync(data)
    }
    closeCategoryForm()
  }

  async function submitItem(data: MenuItemForm) {
    if (!selectedCategoryId) return
    const payload = {
      ...data,
      description: data.description || null,
      imageUrl: data.imageUrl || null,
    }
    if (editingItem) {
      await itemUpdate.mutateAsync({ id: editingItem.id, ...payload })
    } else {
      await itemCreate.mutateAsync({ categoryId: selectedCategoryId, ...payload })
    }
    closeItemForm()
  }

  const categories = catQuery.data ?? []
  const items = itemQuery.data ?? []

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-6xl mx-auto p-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-6">Menu Management</h1>

        <div className="flex gap-6">
          {/* Left panel — categories */}
          <div className="w-64 shrink-0">
            <div className="bg-white rounded-lg border shadow-sm p-4">
              <div className="flex items-center justify-between mb-3">
                <h2 className="font-semibold text-gray-700">Categories</h2>
                <button
                  onClick={() => { closeCategoryForm(); setShowCategoryForm(true) }}
                  className="text-xs bg-amber-500 text-white px-2 py-1 rounded hover:bg-amber-600"
                >
                  + Add
                </button>
              </div>

              {catQuery.isLoading && <p className="text-sm text-gray-400">Loading…</p>}

              <ul className="space-y-1">
                {categories.map((cat) => (
                  <li
                    key={cat.id}
                    className={`flex items-center justify-between rounded px-2 py-1.5 cursor-pointer text-sm ${selectedCategoryId === cat.id ? 'bg-amber-50 text-amber-800 font-medium' : 'hover:bg-gray-50 text-gray-700'}`}
                  >
                    <span onClick={() => setSelectedCategoryId(cat.id)} className="flex-1 truncate">
                      {cat.name}
                    </span>
                    <div className="flex gap-1 ml-1">
                      <button onClick={() => openCategoryEdit(cat)} className="text-gray-400 hover:text-gray-700 text-xs">✎</button>
                      <button
                        onClick={() => { if (confirm(`Delete "${cat.name}"?`)) catDelete.mutate(cat.id) }}
                        className="text-gray-400 hover:text-red-600 text-xs"
                      >
                        ✕
                      </button>
                    </div>
                  </li>
                ))}
              </ul>

              {showCategoryForm && (
                <form onSubmit={catForm.handleSubmit(submitCategory)} className="mt-4 space-y-2 border-t pt-4">
                  <p className="text-xs font-medium text-gray-600">{editingCategory ? 'Edit category' : 'New category'}</p>
                  <div>
                    <input
                      {...catForm.register('name')}
                      placeholder="Name"
                      className="w-full border rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-amber-400"
                    />
                    {catForm.formState.errors.name && <p className="text-red-500 text-xs mt-0.5">{catForm.formState.errors.name.message}</p>}
                  </div>
                  <div>
                    <input
                      {...catForm.register('displayOrder')}
                      type="number"
                      placeholder="Display order"
                      className="w-full border rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-amber-400"
                    />
                    {catForm.formState.errors.displayOrder && <p className="text-red-500 text-xs mt-0.5">{catForm.formState.errors.displayOrder.message}</p>}
                  </div>
                  <div className="flex gap-2">
                    <button type="submit" disabled={catCreate.isPending || catUpdate.isPending} className="flex-1 bg-amber-500 text-white text-xs py-1 rounded hover:bg-amber-600 disabled:opacity-60">
                      Save
                    </button>
                    <button type="button" onClick={closeCategoryForm} className="flex-1 bg-gray-100 text-gray-700 text-xs py-1 rounded hover:bg-gray-200">
                      Cancel
                    </button>
                  </div>
                </form>
              )}
            </div>
          </div>

          {/* Right panel — menu items */}
          <div className="flex-1">
            <div className="bg-white rounded-lg border shadow-sm p-4">
              <div className="flex items-center justify-between mb-4">
                <h2 className="font-semibold text-gray-700">
                  {selectedCategoryId
                    ? `Items — ${categories.find((c) => c.id === selectedCategoryId)?.name ?? ''}`
                    : 'Select a category'}
                </h2>
                {selectedCategoryId && (
                  <button
                    onClick={() => { closeItemForm(); setShowItemForm(true) }}
                    className="text-xs bg-amber-500 text-white px-2 py-1 rounded hover:bg-amber-600"
                  >
                    + Add item
                  </button>
                )}
              </div>

              {!selectedCategoryId && (
                <p className="text-sm text-gray-400">Choose a category on the left to manage its items.</p>
              )}

              {selectedCategoryId && itemQuery.isLoading && <p className="text-sm text-gray-400">Loading…</p>}

              {selectedCategoryId && showItemForm && (
                <form onSubmit={itemForm.handleSubmit(submitItem)} className="mb-4 p-3 bg-gray-50 rounded border space-y-2">
                  <p className="text-xs font-medium text-gray-600">{editingItem ? 'Edit item' : 'New item'}</p>
                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <input {...itemForm.register('name')} placeholder="Name" className="w-full border rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-amber-400" />
                      {itemForm.formState.errors.name && <p className="text-red-500 text-xs">{itemForm.formState.errors.name.message}</p>}
                    </div>
                    <div>
                      <input {...itemForm.register('priceMinor')} type="number" placeholder="Price (IDR minor)" className="w-full border rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-amber-400" />
                      {itemForm.formState.errors.priceMinor && <p className="text-red-500 text-xs">{itemForm.formState.errors.priceMinor.message}</p>}
                    </div>
                  </div>
                  <input {...itemForm.register('description')} placeholder="Description (optional)" className="w-full border rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-amber-400" />
                  <input {...itemForm.register('imageUrl')} placeholder="Image URL (optional)" className="w-full border rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-amber-400" />
                  {itemForm.formState.errors.imageUrl && <p className="text-red-500 text-xs">{itemForm.formState.errors.imageUrl.message}</p>}
                  <label className="flex items-center gap-2 text-sm text-gray-700">
                    <input type="checkbox" {...itemForm.register('isAvailable')} className="rounded" />
                    Available
                  </label>
                  <div className="flex gap-2">
                    <button type="submit" disabled={itemCreate.isPending || itemUpdate.isPending} className="flex-1 bg-amber-500 text-white text-xs py-1 rounded hover:bg-amber-600 disabled:opacity-60">Save</button>
                    <button type="button" onClick={closeItemForm} className="flex-1 bg-gray-100 text-gray-700 text-xs py-1 rounded hover:bg-gray-200">Cancel</button>
                  </div>
                </form>
              )}

              <div className="space-y-2">
                {items.map((item) => (
                  <div key={item.id} className="flex items-center justify-between border rounded px-3 py-2">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium text-gray-800 truncate">{item.name}</span>
                        {!item.isAvailable && (
                          <span className="text-xs bg-red-100 text-red-600 px-1.5 py-0.5 rounded">Unavailable</span>
                        )}
                      </div>
                      {item.description && <p className="text-xs text-gray-400 truncate">{item.description}</p>}
                      <p className="text-xs text-amber-600 font-medium">{formatIDR(item.priceMinor)}</p>
                    </div>
                    <div className="flex gap-2 ml-3">
                      <button
                        onClick={() =>
                          itemUpdate.mutate({ id: item.id, isAvailable: !item.isAvailable })
                        }
                        className={`text-xs px-2 py-1 rounded ${item.isAvailable ? 'bg-green-100 text-green-700 hover:bg-green-200' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}`}
                      >
                        {item.isAvailable ? 'Available' : 'Unavailable'}
                      </button>
                      <button onClick={() => openItemEdit(item)} className="text-gray-400 hover:text-gray-700 text-xs px-1">✎</button>
                      <button
                        onClick={() => { if (confirm(`Delete "${item.name}"?`)) itemDelete.mutate(item.id) }}
                        className="text-gray-400 hover:text-red-600 text-xs px-1"
                      >
                        ✕
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
