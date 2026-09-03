import { AlertCircle, BarChart3, Pencil, Trash2 } from 'lucide-react'
import { useMemo } from 'react'
import { BarChart } from '@/components/BarChart'
import { GaugeChart, type Threshold } from '@/components/GaugeChart'
import { LineChart } from '@/components/LineChart'
import { LogViewer } from '@/components/LogViewer'
import { PieChart, type PieDataItem } from '@/components/PieChart'
import { ensurePanelTypesRegistered } from '@/components/panels/registerPanelTypes'
import { StatPanel } from '@/components/StatPanel'
import { TablePanel } from '@/components/TablePanel'
import { TraceHeatmapPanel } from '@/components/TraceHeatmapPanel'
import { TraceListPanel } from '@/components/TraceListPanel'
import { useCrosshairSync } from '@/contexts/CrosshairSyncContext'
import { useDashboardVariables } from '@/contexts/VariablesContext'
import { usePanelData } from '@/hooks/usePanelData'
import { useTimeRange } from '@/hooks/useTimeRange'
import type { Panel as PanelType } from '@/types/panel'

ensurePanelTypesRegistered()

const WIDE_PANEL_COLS = 8

function panelQueryExpr(query: PanelType['query']): string {
  if (!query) return ''
  if (typeof query.expr === 'string' && query.expr.trim()) return query.expr
  if (typeof query.promql === 'string' && query.promql.trim()) return query.promql
  return ''
}

type PanelProps = {
  panel: PanelType
  onEdit?: (panel: PanelType) => void
  onDelete?: (panel: PanelType) => void
  onOpenTrace?: (payload: { datasourceId: string; traceId: string }) => void
}

export function Panel({ panel, onEdit, onDelete, onOpenTrace }: PanelProps) {
  const { interpolate, variables } = useDashboardVariables()
  const { groupId } = useCrosshairSync()
  const { selectedPreset, isCustomRange } = useTimeRange()
  const variableSignature = useMemo(
    () =>
      variables
        .map((variable) => {
          const current = Array.isArray(variable.current)
            ? variable.current.join(',')
            : (variable.current ?? '')
          return `${variable.name}:${current}`
        })
        .join('|'),
    [variables],
  )
  const { loading, error, chartSeries, logs, traceSummaries, hasQuery, registry } = usePanelData(
    panel,
    interpolate,
    variableSignature,
  )

  const gaugeValue = useMemo(() => {
    if (chartSeries.length === 0) return 0
    const firstSeries = chartSeries[0]
    if (!firstSeries || firstSeries.data.length === 0) return 0
    return firstSeries.data[firstSeries.data.length - 1]!.value
  }, [chartSeries])

  const gaugeConfig = useMemo(() => {
    const query = panel.query || {}
    return {
      min: typeof query.min === 'number' ? query.min : 0,
      max: typeof query.max === 'number' ? query.max : 100,
      unit: typeof query.unit === 'string' ? query.unit : '',
      decimals: typeof query.decimals === 'number' ? query.decimals : 2,
      thresholds: Array.isArray(query.thresholds) ? (query.thresholds as Threshold[]) : [],
    }
  }, [panel.query])

  const pieData = useMemo<PieDataItem[]>(
    () =>
      chartSeries.map((series) => ({
        name: series.name,
        value: series.data.length > 0 ? series.data[series.data.length - 1]!.value : 0,
      })),
    [chartSeries],
  )

  const pieConfig = useMemo(() => {
    const query = panel.query || {}
    return {
      displayAs: (query.displayAs === 'donut' ? 'donut' : 'pie') as 'pie' | 'donut',
      showLegend: query.showLegend !== false,
      showLabels: query.showLabels !== false,
    }
  }, [panel.query])

  const statData = useMemo(
    () =>
      chartSeries[0]?.data.map((point) => ({
        timestamp: point.timestamp,
        value: point.value,
      })) ?? [],
    [chartSeries],
  )

  const statValue = useMemo(() => {
    if (statData.length === 0) return 0
    return statData[statData.length - 1]!.value
  }, [statData])

  const statPreviousValue = useMemo(() => {
    if (statData.length < 2) return undefined
    return statData[statData.length - 2]!.value
  }, [statData])

  const statConfig = useMemo(() => {
    const query = panel.query || {}
    return {
      unit: typeof query.unit === 'string' ? query.unit : '',
      decimals: typeof query.decimals === 'number' ? query.decimals : 2,
      showTrend: query.showTrend !== false,
      showSparkline: query.showSparkline !== false,
      thresholds: Array.isArray(query.thresholds) ? (query.thresholds as Threshold[]) : [],
    }
  }, [panel.query])

  const textContent = typeof panel.query?.content === 'string' ? panel.query.content : ''

  const isLineChart = panel.type === 'line_chart'
  const isBarChart = panel.type === 'bar_chart'
  const isGaugeChart = panel.type === 'gauge'
  const isPieChart = panel.type === 'pie'
  const isStatPanel = panel.type === 'stat'
  const isTablePanel = panel.type === 'table'
  const isLogPanel = panel.type === 'logs'
  const isTraceListPanel = panel.type === 'trace_list'
  const isTraceHeatmapPanel = panel.type === 'trace_heatmap'
  const isTextPanel = panel.type === 'text'
  const isRegistryPanel = registry !== null
  const registryEmptyState = registry?.emptyState ?? null
  const isUnsupportedRegistryPanel = registry?.supportStatus === 'unsupported'
  const isSetupRequiredRegistryPanel = registry?.supportStatus === 'setup_required'
  const isWidePanel = panel.grid_pos.w >= WIDE_PANEL_COLS
  const queryExpr = panelQueryExpr(panel.query)
  const headerMeta = isWidePanel && queryExpr ? queryExpr : isCustomRange ? '' : selectedPreset

  function handleOpenTrace(traceId: string) {
    const datasourceId = panel.query?.datasource_id
    if (typeof datasourceId !== 'string' || !datasourceId) return
    onOpenTrace?.({ datasourceId, traceId })
  }

  function renderRegistryEmptyState() {
    return (
      <div
        data-testid="panel-unsupported-empty"
        className="flex flex-1 flex-col items-center justify-center gap-2 px-6 text-center"
        style={{ color: 'var(--color-on-surface-variant)' }}
      >
        <AlertCircle size={40} style={{ color: 'var(--color-tertiary)' }} />
        <h4 className="m-0 text-sm font-semibold" style={{ color: 'var(--color-on-surface)' }}>
          {registryEmptyState?.title ||
            (isRegistryPanel
              ? `${registry?.label || panel.type} not available in React yet`
              : 'Panel not supported yet')}
        </h4>
        <p className="m-0 max-w-sm text-xs leading-5">
          {registryEmptyState?.description ||
            'This panel type is registered but has no live React renderer yet. Use a Core panel type for live dashboard content.'}
        </p>
        {registryEmptyState?.actionLabel ? (
          <span
            className="text-[11px] font-semibold uppercase tracking-wide"
            style={{ color: 'var(--color-primary)' }}
          >
            {registryEmptyState.actionLabel}
          </span>
        ) : null}
      </div>
    )
  }

  function renderBody() {
    // Text is the only registry type with a live React body; handle it before support gates.
    if (isTextPanel) {
      return (
        <div
          className="min-h-0 flex-1 overflow-auto whitespace-pre-wrap p-3 text-sm leading-6"
          style={{ color: 'var(--color-on-surface)' }}
          data-testid="text-panel"
        >
          {textContent || 'Edit this panel to add content.'}
        </div>
      )
    }

    // Non-live registry types: skip loading/query chrome so the empty state is immediate.
    if (isUnsupportedRegistryPanel || isSetupRequiredRegistryPanel || isRegistryPanel) {
      return renderRegistryEmptyState()
    }

    if (!hasQuery) {
      return (
        <div
          className="flex flex-1 flex-col items-center justify-center gap-3"
          style={{ color: 'var(--color-outline)' }}
        >
          <BarChart3 size={48} />
          <p className="m-0 text-sm">No query configured</p>
          {onEdit ? (
            <button
              type="button"
              className="cursor-pointer rounded-lg border-0 px-4 py-2 text-sm font-medium text-white transition hover:opacity-90"
              style={{
                background:
                  'linear-gradient(135deg, var(--color-primary), var(--color-primary-dim))',
              }}
              data-testid="panel-configure-btn"
              onClick={() => onEdit(panel)}
            >
              Configure Panel
            </button>
          ) : null}
        </div>
      )
    }

    if (loading) {
      return (
        <div className="flex flex-1 flex-col items-center justify-center gap-3">
          <div
            className="h-8 w-8 animate-spin rounded-full border-[3px]"
            style={{
              borderColor: 'var(--color-outline-variant)',
              borderTopColor: 'var(--color-primary)',
            }}
          />
          <p className="m-0 text-sm" style={{ color: 'var(--color-on-surface-variant)' }}>
            Loading data...
          </p>
        </div>
      )
    }

    if (error) {
      return (
        <div className="flex flex-1 flex-col items-center justify-center gap-3">
          <AlertCircle size={48} style={{ color: 'var(--color-error)' }} />
          <p className="m-0 p-2 text-xs" style={{ color: 'var(--color-error)' }}>
            {error}
          </p>
        </div>
      )
    }

    if (isLineChart && chartSeries.length > 0) {
      return (
        <div className="min-h-0 flex-1">
          <LineChart series={chartSeries} group={groupId ?? undefined} />
        </div>
      )
    }

    if (isBarChart && chartSeries.length > 0) {
      return (
        <div className="min-h-0 flex-1">
          <BarChart series={chartSeries} />
        </div>
      )
    }

    if (isGaugeChart && chartSeries.length > 0) {
      return (
        <div className="min-h-0 flex-1">
          <GaugeChart value={gaugeValue} {...gaugeConfig} />
        </div>
      )
    }

    if (isPieChart && pieData.length > 0) {
      return (
        <div className="min-h-0 flex-1">
          <PieChart data={pieData} {...pieConfig} />
        </div>
      )
    }

    if (isStatPanel && statData.length > 0) {
      return (
        <div className="min-h-0 flex-1">
          <StatPanel
            value={statValue}
            previousValue={statPreviousValue}
            data={statData}
            label={panel.title}
            unit={statConfig.unit}
            decimals={statConfig.decimals}
            thresholds={statConfig.thresholds}
            showTrend={statConfig.showTrend}
            showSparkline={statConfig.showSparkline}
          />
        </div>
      )
    }

    if (isTablePanel && chartSeries.length > 0) {
      return (
        <div className="min-h-0 flex-1">
          <TablePanel series={chartSeries} />
        </div>
      )
    }

    if (isLogPanel && logs.length > 0) {
      return (
        <div className="min-h-0 flex-1">
          <LogViewer logs={logs} />
        </div>
      )
    }

    if (isTraceListPanel && traceSummaries.length > 0) {
      return (
        <div className="min-h-0 flex-1">
          <TraceListPanel traces={traceSummaries} onOpenTrace={handleOpenTrace} />
        </div>
      )
    }

    if (isTraceHeatmapPanel && traceSummaries.length > 0) {
      return (
        <div className="min-h-0 flex-1">
          <TraceHeatmapPanel traces={traceSummaries} onOpenTrace={handleOpenTrace} />
        </div>
      )
    }

    return (
      <div
        className="flex flex-1 flex-col items-center justify-center gap-3"
        style={{ color: 'var(--color-on-surface-variant)' }}
      >
        <AlertCircle size={48} style={{ color: 'var(--color-tertiary)' }} />
        <p className="m-0 text-sm">No data available</p>
      </div>
    )
  }

  return (
    <div
      className="relative flex h-full flex-col gap-2 overflow-hidden rounded-lg p-[var(--panel-padding)]"
      style={{
        backgroundColor: 'var(--color-surface-container-low)',
        borderWidth: '1px',
        borderStyle: 'solid',
        borderColor: 'var(--color-stroke-subtle)',
      }}
      data-testid={`dashboard-panel-${panel.id}`}
    >
      <div className="panel-header flex items-center justify-between gap-2">
        <h3
          className={`min-w-0 truncate font-medium ${isWidePanel ? 'text-[13px]' : 'text-xs'}`}
          style={{
            color: isWidePanel ? 'var(--color-on-surface)' : 'var(--color-on-surface-variant)',
          }}
          data-testid="panel-title"
        >
          {panel.title}
        </h3>
        <div className="flex min-w-0 shrink-0 items-center gap-1">
          {headerMeta ? (
            <span
              className={`max-w-[220px] truncate ${isWidePanel ? 'font-mono text-[11px]' : 'text-[11px]'}`}
              style={{ color: 'var(--color-on-surface-variant)' }}
              data-testid="panel-header-meta"
              title={headerMeta}
            >
              {headerMeta}
            </span>
          ) : null}
          {(onEdit || onDelete) && (
            <div className="panel-actions flex gap-1">
              {onEdit ? (
                <button
                  type="button"
                  className="flex h-7 w-7 cursor-pointer items-center justify-center rounded-md border-0 bg-transparent transition hover:opacity-80"
                  style={{ color: 'var(--color-outline)' }}
                  data-testid="panel-edit-btn"
                  title="Edit panel"
                  onClick={() => onEdit(panel)}
                >
                  <Pencil size={16} />
                </button>
              ) : null}
              {onDelete ? (
                <button
                  type="button"
                  className="flex h-7 w-7 cursor-pointer items-center justify-center rounded-md border-0 bg-transparent transition hover:opacity-80"
                  style={{ color: 'var(--color-outline)' }}
                  data-testid="panel-delete-btn"
                  title="Delete panel"
                  onClick={() => onDelete(panel)}
                >
                  <Trash2 size={16} />
                </button>
              ) : null}
            </div>
          )}
        </div>
      </div>

      <div className="flex min-h-0 flex-1 flex-col overflow-hidden">{renderBody()}</div>
    </div>
  )
}
