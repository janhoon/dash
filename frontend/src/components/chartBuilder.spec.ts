import { describe, expect, it } from 'vitest'
import {
  CHART_BUILDER_DEFAULT_QUERY,
  CHART_BUILDER_PIN_LABEL,
  CHART_BUILDER_PREVIEW_BARS,
  CHART_BUILDER_TYPES,
  chartBuilderTypeLabel,
  isChartBuilderType,
  resolveChartBuilderTitle,
} from '@/components/chartBuilder'

describe('chartBuilder', () => {
  it('exposes the Figma 12:73 type chips in order', () => {
    expect(CHART_BUILDER_TYPES.map((entry) => entry.label)).toEqual([
      'Time series',
      'Stat',
      'Table',
      'Heatmap',
    ])
    expect(CHART_BUILDER_TYPES.map((entry) => entry.type)).toEqual([
      'line_chart',
      'stat',
      'table',
      'heatmap',
    ])
  })

  it('keeps the Figma default query, pin copy, and 16 preview bars', () => {
    expect(CHART_BUILDER_DEFAULT_QUERY).toBe('rate(http_requests_total[5m])')
    expect(CHART_BUILDER_PIN_LABEL).toBe('Pin to Victoria / prod')
    expect(CHART_BUILDER_PREVIEW_BARS).toHaveLength(16)
  })

  it('resolves an empty title from the query', () => {
    expect(resolveChartBuilderTitle('', CHART_BUILDER_DEFAULT_QUERY)).toBe(
      CHART_BUILDER_DEFAULT_QUERY,
    )
    expect(resolveChartBuilderTitle('CPU', 'up')).toBe('CPU')
    expect(resolveChartBuilderTitle('  ', '  ')).toBe('Panel')
  })

  it('labels known builder types and falls back to Time series', () => {
    expect(chartBuilderTypeLabel('line_chart')).toBe('Time series')
    expect(chartBuilderTypeLabel('stat')).toBe('Stat')
    expect(isChartBuilderType('line_chart')).toBe(true)
    expect(isChartBuilderType('gauge')).toBe(false)
    expect(chartBuilderTypeLabel('gauge')).toBe('Time series')
  })
})
