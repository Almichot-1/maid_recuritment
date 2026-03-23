import { format, formatDistanceToNow } from "date-fns"

export function formatDateTime(value?: string | null) {
  if (!value) {
    return "N/A"
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  return format(date, "MMM d, yyyy h:mm a")
}

export function formatShortDate(value?: string | null) {
  if (!value) {
    return "N/A"
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  return format(date, "MMM d, yyyy")
}

export function formatRelative(value?: string | null) {
  if (!value) {
    return "N/A"
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  return formatDistanceToNow(date, { addSuffix: true })
}

export function formatPercent(value: number) {
  return `${(value * 100).toFixed(1)}%`
}

export function titleize(value: string) {
  return value
    .replace(/_/g, " ")
    .split(" ")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ")
}

export function downloadCsv(filename: string, rows: Array<Record<string, unknown>>) {
  if (!rows.length || typeof window === "undefined") {
    return
  }

  const headers = Object.keys(rows[0])
  const escapeCell = (value: unknown) =>
    `"${String(value ?? "")
      .replace(/"/g, '""')
      .replace(/\n/g, " ")}"`

  const lines = [
    headers.join(","),
    ...rows.map((row) => headers.map((header) => escapeCell(row[header])).join(",")),
  ]

  const blob = new Blob([lines.join("\n")], { type: "text/csv;charset=utf-8;" })
  const url = URL.createObjectURL(blob)
  const link = document.createElement("a")
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

export function buildDailySeries(items: string[], days: number) {
  const today = new Date()
  today.setHours(0, 0, 0, 0)

  const buckets = Array.from({ length: days }, (_, index) => {
    const bucketDate = new Date(today)
    bucketDate.setDate(today.getDate() - (days - index - 1))
    return {
      label: format(bucketDate, "MMM d"),
      key: bucketDate.toISOString().slice(0, 10),
      value: 0,
    }
  })

  for (const item of items) {
    const date = new Date(item)
    if (Number.isNaN(date.getTime())) {
      continue
    }
    const key = new Date(date.getFullYear(), date.getMonth(), date.getDate()).toISOString().slice(0, 10)
    const bucket = buckets.find((entry) => entry.key === key)
    if (bucket) {
      bucket.value += 1
    }
  }

  return buckets
}
