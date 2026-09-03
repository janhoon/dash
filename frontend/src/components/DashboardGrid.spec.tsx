import { render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { DashboardGrid, GRID_GUTTER } from '@/components/DashboardGrid'
import type { Panel as PanelType } from '@/types/panel'

vi.mock('@/components/Panel', () => ({
  Panel: ({ panel }: { panel: PanelType }) => (
    <div data-testid={`dashboard-panel-${panel.id}`}>{panel.title}</div>
  ),
}))

vi.mock('react-grid-layout/legacy', () => ({
  default: ({ children, margin }: { children: ReactNode; margin?: [number, number] }) => (
    <div data-testid="grid-layout" data-margin={margin?.join(',')}>
      {children}
    </div>
  ),
}))

const panels: PanelType[] = [
  {
    id: 'panel-1',
    dashboard_id: 'dash-1',
    title: 'CPU',
    type: 'line_chart',
    grid_pos: { x: 0, y: 0, w: 6, h: 3 },
    query: { expr: 'up' },
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  },
]

describe('DashboardGrid', () => {
  it('uses the 16px design-system grid gutter', () => {
    expect(GRID_GUTTER).toBe(16)

    render(<DashboardGrid panels={panels} onPanelsChange={() => undefined} />)

    expect(screen.getByTestId('dashboard-grid')).toBeTruthy()
    expect(screen.getByTestId('grid-layout').getAttribute('data-margin')).toBe('16,16')
    expect(screen.getByTestId('dashboard-panel-panel-1')).toBeTruthy()
  })
})
