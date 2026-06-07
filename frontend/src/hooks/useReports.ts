import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '../lib/api'
import type { BestSellerRow, HourlyVolumeRow, RevenueRow } from '../types/api'

export const reportKeys = {
  all: ['reports'] as const,
  revenue: (granularity: string, from: string, to: string) =>
    [...reportKeys.all, 'revenue', { granularity, from, to }] as const,
  bestSellers: (from: string, to: string, limit: number) =>
    [...reportKeys.all, 'best-sellers', { from, to, limit }] as const,
  hourlyVolume: (date: string) => [...reportKeys.all, 'hourly-volume', date] as const,
}

export function useRevenueReport(granularity: string, from: string, to: string) {
  const params = new URLSearchParams({ granularity, from, to })
  return useQuery({
    queryKey: reportKeys.revenue(granularity, from, to),
    queryFn: () =>
      apiFetch<{ rows: RevenueRow[] }>(`/api/reports/revenue?${params}`).then((d) => d.rows),
    enabled: !!(from && to),
  })
}

export function useBestSellersReport(from: string, to: string, limit = 10) {
  const params = new URLSearchParams({ from, to, limit: String(limit) })
  return useQuery({
    queryKey: reportKeys.bestSellers(from, to, limit),
    queryFn: () =>
      apiFetch<{ rows: BestSellerRow[] }>(`/api/reports/best-sellers?${params}`).then(
        (d) => d.rows,
      ),
    enabled: !!(from && to),
  })
}

export function useHourlyVolumeReport(date: string) {
  return useQuery({
    queryKey: reportKeys.hourlyVolume(date),
    queryFn: () =>
      apiFetch<{ rows: HourlyVolumeRow[] }>(`/api/reports/hourly-volume?date=${date}`).then(
        (d) => d.rows,
      ),
    enabled: !!date,
  })
}
