const formatter = new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  minimumFractionDigits: 0,
})

/** Format minor-unit IDR (integer cents) to a human-readable string, e.g. 15000 → "Rp 15.000". */
export function formatIDR(minor: number): string {
  return formatter.format(minor)
}
