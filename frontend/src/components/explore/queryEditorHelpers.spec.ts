import { describe, expect, it } from 'vitest'
import {
  applyMetricSuggestion,
  formatLastRanStatus,
  formatUnknownMetricError,
  getQueryLanguageLabel,
  isUnknownMetricError,
  parseMetricSuggestionFromError,
  suggestMetricCorrection,
} from '@/components/explore/queryEditorHelpers'

describe('queryEditorHelpers', () => {
  it('labels prometheus-like datasources as PromQL', () => {
    expect(getQueryLanguageLabel('prometheus')).toBe('PromQL')
    expect(getQueryLanguageLabel('victoriametrics')).toBe('PromQL')
    expect(getQueryLanguageLabel(undefined)).toBe('PromQL')
  })

  it('labels other metrics languages without inventing a new query language', () => {
    expect(getQueryLanguageLabel('clickhouse')).toBe('SQL')
    expect(getQueryLanguageLabel('cloudwatch')).toBe('CloudWatch')
    expect(getQueryLanguageLabel('elasticsearch')).toBe('Elasticsearch')
  })

  it('formats last-ran status like Refresh 12:2', () => {
    expect(formatLastRanStatus(1_000, 3, 5_000)).toBe('Last ran 4s · 3 series')
    expect(formatLastRanStatus(1_000, 1, 1_000)).toBe('Last ran 0s · 1 series')
    expect(formatLastRanStatus(1_000, 2, 61_000)).toBe('Last ran 1m · 2 series')
  })

  it('suggests the closest known metric for a typo in PromQL', () => {
    const suggestion = suggestMetricCorrection('rate(http_requests_totall[5m])', [
      'up',
      'http_requests_total',
    ])

    expect(suggestion).toEqual({
      unknown: 'http_requests_totall',
      suggestion: 'http_requests_total',
    })
    expect(formatUnknownMetricError(suggestion!)).toBe(
      'Unknown metric http_requests_totall. Did you mean http_requests_total?',
    )
  })

  it('does not suggest when the query already uses known metrics', () => {
    expect(
      suggestMetricCorrection(
        'histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))',
        ['http_request_duration_seconds_bucket', 'up'],
      ),
    ).toBeNull()
  })

  it('applies a metric suggestion inside the existing query', () => {
    expect(
      applyMetricSuggestion('rate(http_requests_totall[5m])', {
        unknown: 'http_requests_totall',
        suggestion: 'http_requests_total',
      }),
    ).toBe('rate(http_requests_total[5m])')
  })

  it('parses a Refresh error banner back into a suggestion', () => {
    expect(
      parseMetricSuggestionFromError(
        'Unknown metric http_requests_totall. Did you mean http_requests_total?',
      ),
    ).toEqual({
      unknown: 'http_requests_totall',
      suggestion: 'http_requests_total',
    })
  })

  it('detects unknown-metric provider errors without treating generic failures as typos', () => {
    expect(isUnknownMetricError('unknown metric name')).toBe(true)
    expect(isUnknownMetricError('parse error: unknown metric name "http_requests_totall"')).toBe(
      true,
    )
    expect(
      isUnknownMetricError(
        'Unknown metric http_requests_totall. Did you mean http_requests_total?',
      ),
    ).toBe(true)
    expect(isUnknownMetricError('Selected datasource did not return metric results')).toBe(false)
    expect(isUnknownMetricError('query timed out')).toBe(false)
    expect(isUnknownMetricError('Query failed')).toBe(false)
  })
})
