import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'
import { PanelChrome, panelHeaderMeta } from '@/components/PanelChrome'
import { useTimeRangeStore } from '@/stores/timeRangeStore'
import type { Panel as PanelType } from '@/types/panel'

const chromePanel: PanelType = {
  id: 'p1',
  dashboard_id: 'd1',
  title: 'Request rate',
  type: 'line_chart',
  grid_pos: { x: 0, y: 0, w: 6, h: 4 },
  query: { expr: 'up' },
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
}

describe('panelHeaderMeta', () => {
  it('shows the query on wide panels', () => {
    expect(
      panelHeaderMeta({
        isWidePanel: true,
        queryExpr: 'rate(http_requests_total[5m])',
        isCustomRange: false,
        selectedPreset: '1h',
      }),
    ).toBe('rate(http_requests_total[5m])')
  })

  it('keeps the query on wide panels even for a custom range', () => {
    expect(
      panelHeaderMeta({
        isWidePanel: true,
        queryExpr: 'up',
        isCustomRange: true,
        selectedPreset: '1h',
      }),
    ).toBe('up')
  })

  it('shows the short range on compact panels', () => {
    expect(
      panelHeaderMeta({
        isWidePanel: false,
        queryExpr: 'up',
        isCustomRange: false,
        selectedPreset: '1h',
      }),
    ).toBe('1h')
  })

  it('hides meta for a custom range on compact panels', () => {
    expect(
      panelHeaderMeta({
        isWidePanel: false,
        queryExpr: 'up',
        isCustomRange: true,
        selectedPreset: '1h',
      }),
    ).toBe('')
  })
})

describe('PanelChrome', () => {
  beforeEach(() => {
    useTimeRangeStore.getState()._reset()
  })

  it('uses surface-card tokens and keeps actions visible', () => {
    render(
      <PanelChrome
        panel={chromePanel}
        onEdit={() => undefined}
        onDelete={() => undefined}
      >
        <div data-testid="live-body" />
      </PanelChrome>,
    )

    const chrome = screen.getByTestId('dashboard-panel-p1')
    expect(chrome.style.backgroundColor).toBe('var(--color-surface-container-low)')
    expect(chrome.style.borderColor).toBe('var(--color-stroke-subtle)')
    expect(screen.getByTestId('live-body')).toBeTruthy()
    expect(screen.getByTestId('panel-header-meta').textContent).toBe('1h')
    expect(screen.getByTestId('panel-edit-btn').parentElement?.className).not.toContain('opacity-0')
  })
})
