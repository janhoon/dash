import { describe, expect, it } from 'vitest'
import { panelQueryExpr } from '@/types/panel'

describe('panelQueryExpr', () => {
  it('prefers non-empty expr over promql', () => {
    expect(panelQueryExpr({ expr: 'cpu', promql: 'up' })).toBe('cpu')
  })

  it('falls back to non-empty promql', () => {
    expect(panelQueryExpr({ promql: 'up' })).toBe('up')
    expect(panelQueryExpr({ expr: '  ', promql: 'node_cpu_seconds_total' })).toBe(
      'node_cpu_seconds_total',
    )
  })

  it('trims and ignores empty strings', () => {
    expect(panelQueryExpr({ expr: '  mem  ' })).toBe('mem')
    expect(panelQueryExpr({ promql: '  up  ' })).toBe('up')
    expect(panelQueryExpr({})).toBe('')
    expect(panelQueryExpr(undefined)).toBe('')
  })
})
