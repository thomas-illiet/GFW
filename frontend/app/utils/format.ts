export function dateTime(value?: string) {
  return value
    ? new Intl.DateTimeFormat('en', {
        dateStyle: 'medium',
        timeStyle: 'short',
      }).format(new Date(value))
    : 'Never'
}
export function relativeTime(value?: string, now = new Date()) {
  if (!value) return 'Never'
  const diff = new Date(value).getTime() - now.getTime()
  const absolute = Math.abs(diff)
  if (absolute < 60_000) return 'just now'

  const units: Array<[Intl.RelativeTimeFormatUnit, number]> = [
    ['year', 365 * 24 * 60 * 60 * 1000],
    ['month', 30 * 24 * 60 * 60 * 1000],
    ['day', 24 * 60 * 60 * 1000],
    ['hour', 60 * 60 * 1000],
    ['minute', 60 * 1000],
  ]
  const [unit, milliseconds] =
    units.find(([, size]) => absolute >= size) ?? units.at(-1)!
  return new Intl.RelativeTimeFormat('en', { numeric: 'always' }).format(
    Math.round(diff / milliseconds),
    unit,
  )
}
export function number(value: number) {
  return new Intl.NumberFormat('en').format(value)
}
export function percent(value: number) {
  return `${value.toFixed(1)}%`
}
