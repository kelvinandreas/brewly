import { useState } from 'react'
import { createRoute } from '@tanstack/react-router'
import { Route as authRoute } from './_auth'
import { useRevenueReport, useBestSellersReport, useHourlyVolumeReport } from '../hooks/useReports'
import { formatIDR } from '../lib/currency'

export const Route = createRoute({
  getParentRoute: () => authRoute,
  path: '/reports',
  component: ReportsPage,
})

function toRFC3339(dateStr: string, endOfDay = false): string {
  return endOfDay ? `${dateStr}T23:59:59Z` : `${dateStr}T00:00:00Z`
}

function today(): string {
  return new Date().toISOString().slice(0, 10)
}

function thirtyDaysAgo(): string {
  const d = new Date()
  d.setDate(d.getDate() - 30)
  return d.toISOString().slice(0, 10)
}

type Tab = 'revenue' | 'best-sellers' | 'hourly'

function ReportsPage() {
  const [tab, setTab] = useState<Tab>('revenue')
  const [from, setFrom] = useState(thirtyDaysAgo())
  const [to, setTo] = useState(today())
  const [granularity, setGranularity] = useState<'day' | 'week' | 'month'>('day')
  const [hourlyDate, setHourlyDate] = useState(today())

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b px-6 py-4">
        <h1 className="text-xl font-bold text-gray-800">Reports</h1>
      </header>

      <div className="p-6 max-w-5xl mx-auto space-y-6">
        {/* Tab bar */}
        <div className="flex gap-1 bg-gray-100 p-1 rounded-lg w-fit">
          {(['revenue', 'best-sellers', 'hourly'] as Tab[]).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`px-4 py-2 text-sm rounded-md font-medium transition-colors ${
                tab === t ? 'bg-white shadow text-gray-900' : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              {t === 'revenue' ? 'Revenue' : t === 'best-sellers' ? 'Best Sellers' : 'Hourly'}
            </button>
          ))}
        </div>

        {tab === 'revenue' && (
          <RevenueTab
            from={from}
            to={to}
            granularity={granularity}
            onFromChange={setFrom}
            onToChange={setTo}
            onGranularityChange={setGranularity}
          />
        )}
        {tab === 'best-sellers' && (
          <BestSellersTab from={from} to={to} onFromChange={setFrom} onToChange={setTo} />
        )}
        {tab === 'hourly' && (
          <HourlyTab date={hourlyDate} onDateChange={setHourlyDate} />
        )}
      </div>
    </div>
  )
}

function DateRangePicker({
  from,
  to,
  onFromChange,
  onToChange,
}: {
  from: string
  to: string
  onFromChange: (v: string) => void
  onToChange: (v: string) => void
}) {
  return (
    <div className="flex items-center gap-3 flex-wrap">
      <label className="flex items-center gap-2 text-sm text-gray-600">
        From
        <input
          type="date"
          value={from}
          onChange={(e) => onFromChange(e.target.value)}
          className="border rounded px-2 py-1 text-sm"
        />
      </label>
      <label className="flex items-center gap-2 text-sm text-gray-600">
        To
        <input
          type="date"
          value={to}
          onChange={(e) => onToChange(e.target.value)}
          className="border rounded px-2 py-1 text-sm"
        />
      </label>
    </div>
  )
}

function RevenueTab({
  from,
  to,
  granularity,
  onFromChange,
  onToChange,
  onGranularityChange,
}: {
  from: string
  to: string
  granularity: 'day' | 'week' | 'month'
  onFromChange: (v: string) => void
  onToChange: (v: string) => void
  onGranularityChange: (v: 'day' | 'week' | 'month') => void
}) {
  const { data: rows, isLoading } = useRevenueReport(
    granularity,
    toRFC3339(from),
    toRFC3339(to, true),
  )

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4 flex-wrap">
        <DateRangePicker from={from} to={to} onFromChange={onFromChange} onToChange={onToChange} />
        <div className="flex gap-1 bg-gray-100 p-1 rounded text-sm">
          {(['day', 'week', 'month'] as const).map((g) => (
            <button
              key={g}
              onClick={() => onGranularityChange(g)}
              className={`px-3 py-1 rounded ${granularity === g ? 'bg-white shadow font-medium' : 'text-gray-500'}`}
            >
              {g}
            </button>
          ))}
        </div>
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b">
            <tr>
              <th className="text-left px-4 py-3 font-medium text-gray-600">Period</th>
              <th className="text-right px-4 py-3 font-medium text-gray-600">Orders</th>
              <th className="text-right px-4 py-3 font-medium text-gray-600">Revenue</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {isLoading && (
              <tr>
                <td colSpan={3} className="text-center py-8 text-gray-400">
                  Loading...
                </td>
              </tr>
            )}
            {!isLoading && (!rows || rows.length === 0) && (
              <tr>
                <td colSpan={3} className="text-center py-8 text-gray-400">
                  No data for this period
                </td>
              </tr>
            )}
            {rows?.map((row, i) => (
              <tr key={i} className="hover:bg-gray-50">
                <td className="px-4 py-3 text-gray-700">{row.period}</td>
                <td className="px-4 py-3 text-right text-gray-600">{row.orderCount}</td>
                <td className="px-4 py-3 text-right font-medium text-gray-800">
                  {formatIDR(row.totalMinor)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function BestSellersTab({
  from,
  to,
  onFromChange,
  onToChange,
}: {
  from: string
  to: string
  onFromChange: (v: string) => void
  onToChange: (v: string) => void
}) {
  const { data: rows, isLoading } = useBestSellersReport(
    toRFC3339(from),
    toRFC3339(to, true),
    10,
  )

  return (
    <div className="space-y-4">
      <DateRangePicker from={from} to={to} onFromChange={onFromChange} onToChange={onToChange} />
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b">
            <tr>
              <th className="text-left px-4 py-3 font-medium text-gray-600">#</th>
              <th className="text-left px-4 py-3 font-medium text-gray-600">Item</th>
              <th className="text-right px-4 py-3 font-medium text-gray-600">Qty</th>
              <th className="text-right px-4 py-3 font-medium text-gray-600">Revenue</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {isLoading && (
              <tr>
                <td colSpan={4} className="text-center py-8 text-gray-400">
                  Loading...
                </td>
              </tr>
            )}
            {!isLoading && (!rows || rows.length === 0) && (
              <tr>
                <td colSpan={4} className="text-center py-8 text-gray-400">
                  No data for this period
                </td>
              </tr>
            )}
            {rows?.map((row, i) => (
              <tr key={row.menuItemId} className="hover:bg-gray-50">
                <td className="px-4 py-3 text-gray-400 font-medium">{i + 1}</td>
                <td className="px-4 py-3 text-gray-800 font-medium">{row.name}</td>
                <td className="px-4 py-3 text-right text-gray-600">{row.totalQuantity}</td>
                <td className="px-4 py-3 text-right font-medium text-gray-800">
                  {formatIDR(row.totalMinor)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function HourlyTab({ date, onDateChange }: { date: string; onDateChange: (v: string) => void }) {
  const { data: rows, isLoading } = useHourlyVolumeReport(date)
  const maxOrders = rows ? Math.max(...rows.map((r) => r.orderCount), 1) : 1

  return (
    <div className="space-y-4">
      <label className="flex items-center gap-2 text-sm text-gray-600">
        Date
        <input
          type="date"
          value={date}
          onChange={(e) => onDateChange(e.target.value)}
          className="border rounded px-2 py-1 text-sm"
        />
      </label>

      <div className="bg-white rounded-lg shadow p-4">
        {isLoading && <p className="text-center text-gray-400 py-8">Loading...</p>}
        {!isLoading && (!rows || rows.length === 0) && (
          <p className="text-center text-gray-400 py-8">No data for this date</p>
        )}
        {rows && rows.length > 0 && (
          <div className="space-y-2">
            {rows.map((row) => (
              <div key={row.hour} className="flex items-center gap-3">
                <span className="text-xs text-gray-500 w-10 text-right">
                  {String(row.hour).padStart(2, '0')}:00
                </span>
                <div className="flex-1 bg-gray-100 rounded-full h-6 overflow-hidden">
                  <div
                    className="bg-amber-400 h-full rounded-full transition-all"
                    style={{ width: `${(row.orderCount / maxOrders) * 100}%` }}
                  />
                </div>
                <span className="text-xs text-gray-600 w-16 text-right">
                  {row.orderCount} orders
                </span>
                <span className="text-xs font-medium text-gray-800 w-24 text-right">
                  {formatIDR(row.totalMinor)}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
