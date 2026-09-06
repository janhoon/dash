import {
  CHART_BUILDER_CAPTION,
  CHART_BUILDER_PREVIEW_BARS,
  CHART_BUILDER_RANGE_LABEL,
} from '@/components/chartBuilder'

type ChartBuilderPreviewProps = {
  typeLabel: string
}

export function ChartBuilderPreview({ typeLabel }: ChartBuilderPreviewProps) {
  return (
    <section
      className="flex h-full min-w-0 flex-1 flex-col gap-3 overflow-hidden rounded-lg p-4"
      style={{
        backgroundColor: 'var(--color-surface-container-low)',
        borderWidth: '1px',
        borderStyle: 'solid',
        borderColor: 'var(--color-surface-container-high)',
      }}
      data-testid="chart-builder-preview"
    >
      <div className="flex w-full items-start justify-between">
        <p className="m-0 text-[13px] font-medium" style={{ color: 'var(--color-on-surface)' }}>
          Preview
        </p>
        <p
          className="m-0 text-[12px]"
          style={{ color: 'var(--color-on-surface-variant)' }}
          data-testid="chart-builder-preview-meta"
        >
          {typeLabel} · {CHART_BUILDER_RANGE_LABEL}
        </p>
      </div>
      <div
        className="flex min-h-0 w-full flex-1 items-end gap-[3px] overflow-hidden"
        data-testid="chart-builder-preview-bars"
      >
        {CHART_BUILDER_PREVIEW_BARS.map((bar) => (
          <div
            key={bar.id}
            className="min-w-0 flex-1 rounded-[2px]"
            style={{
              height: `${bar.height}px`,
              backgroundColor: 'var(--color-viz-0)',
            }}
          />
        ))}
      </div>
      <p
        className="m-0 text-[12px]"
        style={{ color: 'var(--color-on-surface-variant)' }}
        data-testid="chart-builder-preview-caption"
      >
        {CHART_BUILDER_CAPTION}
      </p>
    </section>
  )
}
