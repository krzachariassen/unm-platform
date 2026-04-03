/**
 * Returns `"1 team"` or `"3 teams"` — handles singular/plural for display tags.
 * If `plural` is omitted, appends "s" to `singular`.
 */
export function pl(count: number, singular: string, plural?: string): string {
  return `${count} ${count === 1 ? singular : (plural ?? singular + 's')}`
}
