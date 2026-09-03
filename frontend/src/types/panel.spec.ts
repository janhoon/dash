import { describe, expect, it } from 'vitest'
import { readPanelQueryExpr } from '@/types/panel'

describe('readPanelQueryExpr', () => {
  it('prefers promql over expr', () => {
    expect(readPanelQueryExpr({ promql: 'up', expr: 'cpu' })).toBe('up')
  })

  it('falls back to expr', () => {
    expect(readPanelQueryExpr({ expr: 'rate(http_requests_total[5m])' })).toBe(
      'rate(http_requests_total[5m])',
    )
  })

  it('trims and ignores empty strings', () => {
    expect(readPanelQueryExpr({ promql: '  ', expr: '  mem  ' })).toBe('mem')
    expect(readPanelQueryExpr({})).toBe('')
    expect(readPanelQueryExpr(undefined)).toBe('')
  })
})
