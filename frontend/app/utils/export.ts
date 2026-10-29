export type CsvValue = string | number | boolean | null | undefined

function escapeCsv(value: CsvValue) {
  const text = value == null ? '' : String(value)
  return /[",\n\r]/.test(text) ? `"${text.replaceAll('"', '""')}"` : text
}

export function exportCsv(
  filename: string,
  columns: string[],
  rows: CsvValue[][],
) {
  const content = toCsv(columns, rows)
  const blob = new Blob([`\uFEFF${content}`], {
    type: 'text/csv;charset=utf-8',
  })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

export function toCsv(columns: string[], rows: CsvValue[][]) {
  return [columns, ...rows]
    .map((row) => row.map(escapeCsv).join(','))
    .join('\r\n')
}
