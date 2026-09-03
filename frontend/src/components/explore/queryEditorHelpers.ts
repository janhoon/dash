import { PROMQL_FUNCTIONS, PROMQL_KEYWORDS } from '@/promql/language'
import type { DataSourceType } from '@/types/datasource'

export type MetricSuggestion = {
  unknown: string
  suggestion: string
}

export const QUERY_LOADING_BAR_HEIGHTS = [
  80, 110, 90, 140, 120, 160, 130, 180, 150, 200, 170, 140, 160, 190, 170, 150,
] as const

const PROMQL_METRIC_TOKEN = /[a-zA-Z_:][a-zA-Z0-9_:]*/g
const UNKNOWN_METRIC_ERROR = /^Unknown metric (\S+)\. Did you mean ([^\s.?]+)\??\.?$/

const SKIP_IDENTIFIERS = new Set([
  ...Object.keys(PROMQL_FUNCTIONS),
  ...PROMQL_KEYWORDS,
  'inf',
  'nan',
])

export function getQueryLanguageLabel(type_: DataSourceType | undefined): string {
  switch (type_) {
    case 'clickhouse':
      return 'SQL'
    case 'cloudwatch':
      return 'CloudWatch'
    case 'elasticsearch':
      return 'Elasticsearch'
    default:
      return 'PromQL'
  }
}

export function formatLastRanStatus(
  lastRanAt: number,
  seriesCount: number,
  now = Date.now(),
): string {
  const elapsedMs = Math.max(0, now - lastRanAt)
  const elapsedSeconds = Math.floor(elapsedMs / 1000)
  const ran = elapsedSeconds < 60 ? `${elapsedSeconds}s` : `${Math.floor(elapsedSeconds / 60)}m`
  const seriesLabel = seriesCount === 1 ? '1 series' : `${seriesCount} series`
  return `Last ran ${ran} · ${seriesLabel}`
}

export function formatUnknownMetricError(suggestion: MetricSuggestion): string {
  return `Unknown metric ${suggestion.unknown}. Did you mean ${suggestion.suggestion}?`
}

export function parseMetricSuggestionFromError(error: string): MetricSuggestion | null {
  const match = UNKNOWN_METRIC_ERROR.exec(error.trim())
  if (!match?.[1] || !match[2]) return null
  return { unknown: match[1], suggestion: match[2] }
}

export function applyMetricSuggestion(query: string, suggestion: MetricSuggestion): string {
  const escaped = suggestion.unknown.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const token = new RegExp(`(?<![A-Za-z0-9_:])${escaped}(?![A-Za-z0-9_:])`)
  return query.replace(token, suggestion.suggestion)
}

export function suggestMetricCorrection(
  query: string,
  metricNames: string[],
): MetricSuggestion | null {
  if (!query.trim() || metricNames.length === 0) return null

  const known = new Set(metricNames)
  const identifiers = extractPromqlIdentifiers(query).filter(
    (token) => !known.has(token) && !SKIP_IDENTIFIERS.has(token),
  )

  let best: { unknown: string; suggestion: string; distance: number } | null = null

  for (const unknown of identifiers) {
    const match = closestMetric(unknown, metricNames)
    if (!match) continue
    if (!best || match.distance < best.distance) {
      best = { unknown, suggestion: match.name, distance: match.distance }
    }
  }

  if (!best) return null
  const maxDistance = Math.max(1, Math.min(2, Math.floor(best.unknown.length * 0.3)))
  if (best.distance > maxDistance) return null
  return { unknown: best.unknown, suggestion: best.suggestion }
}

function extractPromqlIdentifiers(query: string): string[] {
  const withoutSelectors = query.replace(/\{[^}]*\}/g, ' ')
  const withoutStrings = withoutSelectors.replace(/"(?:[^"\\]|\\.)*"/g, ' ')
  const tokens = withoutStrings.match(PROMQL_METRIC_TOKEN) ?? []
  return [...new Set(tokens)]
}

function closestMetric(
  unknown: string,
  metricNames: string[],
): { name: string; distance: number } | null {
  let best: { name: string; distance: number } | null = null
  const unknownLower = unknown.toLowerCase()

  for (const name of metricNames) {
    if (Math.abs(name.length - unknown.length) > 2) continue
    const distance = levenshtein(unknownLower, name.toLowerCase())
    if (!best || distance < best.distance) {
      best = { name, distance }
    }
  }

  return best
}

function levenshtein(a: string, b: string): number {
  if (a === b) return 0
  if (a.length === 0) return b.length
  if (b.length === 0) return a.length

  const prev = Array.from({ length: b.length + 1 }, (_, index) => index)
  const next = new Array<number>(b.length + 1)

  for (let i = 1; i <= a.length; i++) {
    next[0] = i
    for (let j = 1; j <= b.length; j++) {
      const cost = a[i - 1] === b[j - 1] ? 0 : 1
      next[j] = Math.min((prev[j] ?? 0) + 1, (next[j - 1] ?? 0) + 1, (prev[j - 1] ?? 0) + cost)
    }
    for (let j = 0; j <= b.length; j++) {
      prev[j] = next[j] ?? 0
    }
  }

  return prev[b.length] ?? b.length
}
