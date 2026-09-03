import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { PanelChrome, panelHeaderMeta, WIDE_PANEL_COLS } from '@/components/PanelChrome'

describe('panelHeaderMeta', () => {
  it('shows the query on wide panels', () => {
    expect(
      panelHeaderMeta({
        gridWidth: WIDE_PANEL_COLS,
        queryExpr: 'rate(http_requests_total[5m])',
        selectedPreset: '1h',
        isCustomRange: false,
      }),
    ).toBe('rate(http_requests_total[5m])')
  })

  it('shows the short range on compact panels', () => {
    expect(
      panelHeaderMeta({
        gridWidth: 4,
        queryExpr: 'up',
        selectedPreset: '1h',
        isCustomRange: false,
      }),
    ).toBe('1h')
  })

  it('hides meta for a custom range on compact panels', () => {
    expect(
      panelHeaderMeta({
        gridWidth: 4,
        queryExpr: 'up',
        selectedPreset: '1h',
        isCustomRange: true,
      }),
    ).toBe('')
  })
})

describe('PanelChrome', () => {
  it('uses surface-card tokens and keeps actions visible', () => {
    render(
      <PanelChrome
        panelId="p1"
        title="Request rate"
        wide={false}
        headerMeta="1h"
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
    expect(screen.getByTestId('panel-edit-btn').parentElement?.className).not.toContain('opacity-0')
  })
})
