export const CHART_BUILDER_DEFAULT_QUERY = 'rate(http_requests_total[5m])'
export const CHART_BUILDER_PIN_LABEL = 'Pin to Victoria / prod'
export const CHART_BUILDER_PIN_FALLBACK_LABEL = 'Pin panel'
export const CHART_BUILDER_SUBTITLE = 'One query. One chart. Then pin.'
export const CHART_BUILDER_CAPTION = 'No decoration. The chart is the panel.'
export const CHART_BUILDER_RANGE_LABEL = 'last 2 hours'

export const CHART_BUILDER_TYPES = [
  { type: 'line_chart', label: 'Time series' },
  { type: 'stat', label: 'Stat' },
  { type: 'table', label: 'Table' },
  { type: 'heatmap', label: 'Heatmap' },
] as const

export type ChartBuilderType = (typeof CHART_BUILDER_TYPES)[number]['type']

const CHART_BUILDER_TYPE_SET = new Set<string>(CHART_BUILDER_TYPES.map((entry) => entry.type))

const CORE_TYPE_LABELS: Record<string, string> = {
  bar_chart: 'Bar Chart',
  pie: 'Pie Chart',
  gauge: 'Gauge',
  logs: 'Logs',
  trace_list: 'Trace List',
  trace_heatmap: 'Trace Heatmap',
  text: 'Text',
}

export const CHART_BUILDER_PREVIEW_BARS = [
  { id: 'b0', height: 165 },
  { id: 'b1', height: 185 },
  { id: 'b2', height: 206 },
  { id: 'b3', height: 198 },
  { id: 'b4', height: 255 },
  { id: 'b5', height: 288 },
  { id: 'b6', height: 272 },
  { id: 'b7', height: 330 },
  { id: 'b8', height: 305 },
  { id: 'b9', height: 363 },
  { id: 'b10', height: 338 },
  { id: 'b11', height: 288 },
  { id: 'b12', height: 313 },
  { id: 'b13', height: 371 },
  { id: 'b14', height: 346 },
  { id: 'b15', height: 321 },
] as const

export type ChartBuilderDatasourceRef = {
  name: string
  type: string
}

export function isChartBuilderType(panelType: string): panelType is ChartBuilderType {
  return CHART_BUILDER_TYPE_SET.has(panelType)
}

export function chartBuilderTypeLabel(panelType: string): string {
  const match = CHART_BUILDER_TYPES.find((entry) => entry.type === panelType)
  if (match) return match.label
  return CORE_TYPE_LABELS[panelType] ?? panelType
}

export function resolveChartBuilderTitle(title: string, queryText: string): string {
  const explicit = title.trim()
  if (explicit) return explicit
  const fromQuery = queryText.trim()
  if (fromQuery) return fromQuery
  return 'Panel'
}

export function chartBuilderPinLabel(datasource: ChartBuilderDatasourceRef | null): string {
  if (!datasource) return CHART_BUILDER_PIN_FALLBACK_LABEL
  const name = datasource.name.trim()
  if (!name) return CHART_BUILDER_PIN_FALLBACK_LABEL
  const blob = `${name} ${datasource.type}`.toLowerCase()
  const isVictoria = blob.includes('victoria')
  const isProd = /\bprod\b/.test(blob)
  if (isVictoria && isProd) return CHART_BUILDER_PIN_LABEL
  return `Pin to ${name}`
}
