import { describe, expect, it } from 'vitest'
import {
  CHART_BUILDER_DEFAULT_QUERY,
  CHART_BUILDER_PIN_FALLBACK_LABEL,
  CHART_BUILDER_PIN_LABEL,
  CHART_BUILDER_PREVIEW_BARS,
  CHART_BUILDER_TYPES,
  chartBuilderPinLabel,
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

  it('labels known builder types without pretending unknown types are Time series', () => {
    expect(chartBuilderTypeLabel('line_chart')).toBe('Time series')
    expect(chartBuilderTypeLabel('stat')).toBe('Stat')
    expect(isChartBuilderType('line_chart')).toBe(true)
    expect(isChartBuilderType('gauge')).toBe(false)
    expect(chartBuilderTypeLabel('gauge')).toBe('Gauge')
    expect(chartBuilderTypeLabel('logs')).toBe('Logs')
    expect(chartBuilderTypeLabel('trace_list')).toBe('Trace List')
    expect(chartBuilderTypeLabel('custom_widget')).toBe('custom_widget')
  })

  it('derives pin copy from the selected datasource', () => {
    expect(chartBuilderPinLabel(null)).toBe(CHART_BUILDER_PIN_FALLBACK_LABEL)
    expect(chartBuilderPinLabel({ name: 'Prometheus Main', type: 'prometheus' })).toBe(
      'Pin to Prometheus Main',
    )
    expect(chartBuilderPinLabel({ name: 'Victoria / prod', type: 'victoriametrics' })).toBe(
      CHART_BUILDER_PIN_LABEL,
    )
    expect(chartBuilderPinLabel({ name: 'prod', type: 'victoriametrics' })).toBe(
      CHART_BUILDER_PIN_LABEL,
    )
    expect(chartBuilderPinLabel({ name: 'VictoriaLogs Main', type: 'victorialogs' })).toBe(
      'Pin to VictoriaLogs Main',
    )
  })
})
