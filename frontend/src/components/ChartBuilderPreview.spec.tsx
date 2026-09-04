import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { ChartBuilderPreview } from '@/components/ChartBuilderPreview'
import { CHART_BUILDER_CAPTION, CHART_BUILDER_PREVIEW_BARS } from '@/components/chartBuilder'

describe('ChartBuilderPreview', () => {
  it('renders the Figma 12:73 preview chrome and 16 accent bars', () => {
    render(<ChartBuilderPreview typeLabel="Time series" />)

    expect(screen.getByTestId('chart-builder-preview')).toBeTruthy()
    expect(screen.getByText('Preview')).toBeTruthy()
    expect(screen.getByTestId('chart-builder-preview-meta').textContent).toBe(
      'Time series · last 2 hours',
    )
    expect(screen.getByTestId('chart-builder-preview-caption').textContent).toBe(
      CHART_BUILDER_CAPTION,
    )

    const bars = screen.getByTestId('chart-builder-preview-bars')
    expect(bars.children).toHaveLength(CHART_BUILDER_PREVIEW_BARS.length)
    const first = bars.children[0] as HTMLElement
    expect(first.style.backgroundColor).toBe('var(--color-primary)')
    expect(first.style.height).toBe('165px')
  })
})
