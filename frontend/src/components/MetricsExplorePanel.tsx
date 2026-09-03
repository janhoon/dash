import { Check, ChevronDown, ChevronUp, History, Star, X } from 'lucide-react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useSearchParams } from 'react-router'
import { fetchDataSourceMetricNames, queryDataSource } from '@/api/datasources'
import { ExportToDashboardModal } from '@/components/ExportToDashboardModal'
import { LineChart, type ChartSeries } from '@/components/LineChart'
import { MonacoQueryEditor } from '@/components/MonacoQueryEditor'
import { TimeRangePicker } from '@/components/TimeRangePicker'
import { Button } from '@/components/ui/button'
import { useMetricsDatasources } from '@/hooks/useMetricsDatasources'
import { useTimeRange } from '@/hooks/useTimeRange'
import {
  type PrometheusQueryData,
  type PrometheusQueryResult,
  transformToChartData,
} from '@/promql/client'
import { useFavoritesStore } from '@/stores/favoritesStore'
import { useOrgStore } from '@/stores/orgStore'
import { dataSourceTypeLabels, type DataSource } from '@/types/datasource'

import {
  type ExploreDatasourceChanged,
  getTypeLogo,
  pushQueryHistory,
  readQueryHistory,
  TRACE_NAVIGATION_MAX_AGE_MS,
} from '@/components/explore/exploreShared'
import {
  type TraceMetricsNavigationContext,
  buildServiceMetricsQuery,
  getDefaultMetricsQuery,
  isPrometheusLike,
  METRICS_HISTORY_KEY,
  TRACE_METRICS_NAVIGATION_CONTEXT_KEY,
} from '@/components/explore/metricsExploreHelpers'
import {
  applyMetricSuggestion,
  formatLastRanStatus,
  formatUnknownMetricError,
  getQueryLanguageLabel,
  isUnknownMetricError,
  type MetricSuggestion,
  suggestMetricCorrection,
} from '@/components/explore/queryEditorHelpers'

type MetricsExplorePanelProps = {
  onDatasourceChanged?: (payload: ExploreDatasourceChanged) => void
}

const QUERY_LOADING_BAR_HEIGHTS = [
  80, 110, 90, 140, 120, 160, 130, 180, 150, 200, 170, 140, 160, 190, 170, 150,
] as const

function LastRanStatus({ lastRanAt, seriesCount }: { lastRanAt: number; seriesCount: number }) {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 1000)
    return () => window.clearInterval(timer)
  }, [])

  return (
    <span
      className="text-[11px] text-[var(--color-on-surface-variant)]"
      data-testid="explore-query-status"
    >
      {formatLastRanStatus(lastRanAt, seriesCount, now)}
    </span>
  )
}

function QueryLoadingSkeleton() {
  return (
    <div className="flex flex-1 flex-col gap-3 p-4" data-testid="explore-query-loading">
      <div className="absolute inset-x-0 top-0 h-px overflow-hidden bg-[var(--color-surface-container-high)]">
        <div className="animate-explore-query-progress h-full w-1/3 bg-[var(--color-on-surface)]" />
      </div>
      <div className="h-3 w-40 rounded-sm bg-[var(--color-surface-bright)] animate-shimmer" />
      <div className="flex flex-1 items-end gap-2 overflow-hidden">
        {QUERY_LOADING_BAR_HEIGHTS.map((height, index) => (
          <div
            // biome-ignore lint/suspicious/noArrayIndexKey: static skeleton bars
            key={index}
            className="w-14 shrink-0 rounded-sm bg-[var(--color-surface-bright)] animate-shimmer"
            style={{ height: `${height}px` }}
          />
        ))}
      </div>
    </div>
  )
}

function QueryEmptyResult() {
  return (
    <div className="flex flex-1 flex-col p-6" data-testid="explore-query-empty">
      <p className="m-0 text-[13px] text-[var(--color-on-surface-variant)]">
        No series. Fix the query to see a chart.
      </p>
    </div>
  )
}

export function MetricsExplorePanel({ onDatasourceChanged }: MetricsExplorePanelProps) {
  const currentOrgId = useOrgStore((state) => state.currentOrgId)
  const { metricsDatasources } = useMetricsDatasources(currentOrgId)
  const { timeRange, onRefresh, setCustomRange } = useTimeRange()
  const toggleFavorite = useFavoritesStore((state) => state.toggleFavorite)
  const isFavorite = useFavoritesStore((state) => state.isFavorite)
  const [searchParams] = useSearchParams()

  const [selectedDatasourceId, setSelectedDatasourceId] = useState('')
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<PrometheusQueryResult | null>(null)
  const [chartSeries, setChartSeries] = useState<ChartSeries[]>([])
  const [showExportModal, setShowExportModal] = useState(false)
  const [queryHistory, setQueryHistory] = useState<string[]>([])
  const [showHistory, setShowHistory] = useState(false)
  const [showDatasourceMenu, setShowDatasourceMenu] = useState(false)
  const [pendingServiceName, setPendingServiceName] = useState('')
  const [pendingStartMs, setPendingStartMs] = useState<number | null>(null)
  const [pendingEndMs, setPendingEndMs] = useState<number | null>(null)
  const [metricNames, setMetricNames] = useState<string[]>([])
  const [metricSuggestion, setMetricSuggestion] = useState<MetricSuggestion | null>(null)
  const [lastRanAt, setLastRanAt] = useState<number | null>(null)

  const datasourceMenuRef = useRef<HTMLDivElement | null>(null)
  const pendingNavigationRef = useRef({
    serviceName: '',
    startMs: null as number | null,
    endMs: null as number | null,
  })
  const queryGenerationRef = useRef(0)
  const selectedDatasourceIdRef = useRef(selectedDatasourceId)
  const previousDatasourceIdRef = useRef(selectedDatasourceId)
  selectedDatasourceIdRef.current = selectedDatasourceId

  const activeDatasource = useMemo(
    () => metricsDatasources.find((ds) => ds.id === selectedDatasourceId) ?? null,
    [metricsDatasources, selectedDatasourceId],
  )

  const hasMetricsDatasources = metricsDatasources.length > 0
  const hasResults = result?.status === 'success' && chartSeries.length > 0
  const seriesCount = chartSeries.length
  const queryLanguageLabel = getQueryLanguageLabel(activeDatasource?.type)
  const compactQueryEditor = isPrometheusLike(activeDatasource?.type ?? 'prometheus')

  const addToHistory = useCallback((q: string) => {
    setQueryHistory((prev) => pushQueryHistory(METRICS_HISTORY_KEY, prev, q))
  }, [])

  const applyQueryError = useCallback(
    (message: string, executedQuery: string) => {
      const suggestion =
        isUnknownMetricError(message) && isPrometheusLike(activeDatasource?.type ?? 'prometheus')
          ? suggestMetricCorrection(executedQuery, metricNames)
          : null

      if (suggestion) {
        setMetricSuggestion(suggestion)
        setError(formatUnknownMetricError(suggestion))
        return
      }

      setMetricSuggestion(null)
      setError(message)
    },
    [activeDatasource?.type, metricNames],
  )

  const runQuery = useCallback(async () => {
    if (!selectedDatasourceId) {
      queryGenerationRef.current += 1
      setLoading(false)
      setError('Select a metrics datasource')
      return
    }

    if (!query.trim()) {
      queryGenerationRef.current += 1
      setLoading(false)
      setError('Query is required')
      return
    }

    const generation = ++queryGenerationRef.current
    const executedQuery = query
    const requestedDatasourceId = selectedDatasourceId
    const isCurrent = () =>
      queryGenerationRef.current === generation &&
      selectedDatasourceIdRef.current === requestedDatasourceId

    setLoading(true)
    setError(null)
    setMetricSuggestion(null)
    setResult(null)
    setChartSeries([])

    try {
      const start = Math.floor(timeRange.start / 1000)
      const end = Math.floor(timeRange.end / 1000)
      const duration = end - start
      const step = Math.max(15, Math.floor(duration / 200))
      const dsType = activeDatasource?.type

      const response = await queryDataSource(requestedDatasourceId, {
        query: executedQuery,
        signal:
          dsType === 'clickhouse' || dsType === 'cloudwatch' || dsType === 'elasticsearch'
            ? 'metrics'
            : undefined,
        start,
        end,
        step,
      })

      if (!isCurrent()) return

      if (response.status === 'error') {
        applyQueryError(response.error || 'Query failed', executedQuery)
      } else if (response.resultType !== 'metrics') {
        setMetricSuggestion(null)
        setError('Selected datasource did not return metric results')
      } else {
        const metricsResponse: PrometheusQueryResult = {
          status: response.status,
          data: response.data as PrometheusQueryData | undefined,
          error: response.error,
        }

        setResult(metricsResponse)
        const chartData = transformToChartData(metricsResponse)
        setChartSeries(
          chartData.series.map((series) => ({
            name: series.name,
            data: series.data,
          })),
        )
        setLastRanAt(Date.now())
        addToHistory(executedQuery)
      }
    } catch (e) {
      if (!isCurrent()) return
      applyQueryError(e instanceof Error ? e.message : 'Failed to execute query', executedQuery)
    } finally {
      if (isCurrent()) {
        setLoading(false)
      }
    }
  }, [
    activeDatasource?.type,
    addToHistory,
    applyQueryError,
    query,
    selectedDatasourceId,
    timeRange.end,
    timeRange.start,
  ])

  const applyTraceMetricsNavigationContext = useCallback(() => {
    const pending = pendingNavigationRef.current
    if (!pending.serviceName && pending.startMs === null && pending.endMs === null) {
      return
    }

    setQuery(buildServiceMetricsQuery(activeDatasource?.type || 'prometheus', pending.serviceName))

    if (pending.startMs !== null && pending.endMs !== null) {
      setCustomRange(pending.startMs, pending.endMs)
    }

    pendingNavigationRef.current = { serviceName: '', startMs: null, endMs: null }
    setPendingServiceName('')
    setPendingStartMs(null)
    setPendingEndMs(null)
  }, [activeDatasource?.type, setCustomRange])

  useEffect(() => {
    const urlQuery = searchParams.get('q')
    if (urlQuery) {
      setQuery(urlQuery)
    }

    try {
      const rawContext = localStorage.getItem(TRACE_METRICS_NAVIGATION_CONTEXT_KEY)
      localStorage.removeItem(TRACE_METRICS_NAVIGATION_CONTEXT_KEY)
      if (rawContext) {
        const parsed = JSON.parse(rawContext) as TraceMetricsNavigationContext
        if (typeof parsed.createdAt === 'number') {
          const ageMs = Date.now() - parsed.createdAt
          if (ageMs <= TRACE_NAVIGATION_MAX_AGE_MS) {
            const serviceName =
              typeof parsed.serviceName === 'string' ? parsed.serviceName.trim() : ''
            const startMs =
              typeof parsed.startMs === 'number' &&
              typeof parsed.endMs === 'number' &&
              parsed.endMs > parsed.startMs
                ? parsed.startMs
                : null
            const endMs =
              typeof parsed.startMs === 'number' &&
              typeof parsed.endMs === 'number' &&
              parsed.endMs > parsed.startMs
                ? parsed.endMs
                : null

            pendingNavigationRef.current = { serviceName, startMs, endMs }
            setPendingServiceName(serviceName)
            setPendingStartMs(startMs)
            setPendingEndMs(endMs)
          }
        }
      }
    } catch {
      // Ignore malformed navigation context.
    }

    setQueryHistory(readQueryHistory(METRICS_HISTORY_KEY))
  }, [searchParams])

  useEffect(() => {
    return () => {
      queryGenerationRef.current += 1
    }
  }, [])

  useEffect(() => {
    const previousId = previousDatasourceIdRef.current
    previousDatasourceIdRef.current = selectedDatasourceId
    if (previousId && previousId !== selectedDatasourceId) {
      queryGenerationRef.current += 1
      setLoading(false)
      setResult(null)
      setChartSeries([])
      setError(null)
      setMetricSuggestion(null)
      setLastRanAt(null)
    }
  }, [selectedDatasourceId])

  useEffect(() => {
    if (!activeDatasource || !isPrometheusLike(activeDatasource.type)) {
      setMetricNames([])
      return
    }

    let cancelled = false
    void fetchDataSourceMetricNames(activeDatasource.id)
      .then((names) => {
        if (!cancelled) setMetricNames(names)
      })
      .catch(() => {
        if (!cancelled) setMetricNames([])
      })

    return () => {
      cancelled = true
    }
  }, [activeDatasource])

  const previousOrgIdRef = useRef<string | null>(null)
  useEffect(() => {
    if (!currentOrgId) return
    if (previousOrgIdRef.current && previousOrgIdRef.current !== currentOrgId) {
      setSelectedDatasourceId('')
      setQuery('')
      setResult(null)
      setChartSeries([])
      setError(null)
      setLastRanAt(null)
      setMetricSuggestion(null)
    }
    previousOrgIdRef.current = currentOrgId
  }, [currentOrgId])

  useEffect(() => {
    if (metricsDatasources.length === 0) {
      setSelectedDatasourceId('')
      return
    }

    const hasSelected = metricsDatasources.some((ds) => ds.id === selectedDatasourceId)
    if (!hasSelected) {
      const defaultDatasource = metricsDatasources.find((ds) => ds.is_default)
      const selected = defaultDatasource || metricsDatasources[0]
      if (!selected) return
      setSelectedDatasourceId(selected.id)

      if (!query.trim()) {
        const defaultQuery = getDefaultMetricsQuery(selected.type)
        if (defaultQuery) {
          setQuery(defaultQuery)
        }
      }
    }
  }, [metricsDatasources, query, selectedDatasourceId])

  useEffect(() => {
    if (!activeDatasource) return

    if (
      pendingServiceName ||
      pendingStartMs !== null ||
      pendingEndMs !== null ||
      pendingNavigationRef.current.serviceName ||
      pendingNavigationRef.current.startMs !== null
    ) {
      applyTraceMetricsNavigationContext()
    }
  }, [
    activeDatasource,
    applyTraceMetricsNavigationContext,
    pendingEndMs,
    pendingServiceName,
    pendingStartMs,
  ])

  // biome-ignore lint/correctness/useExhaustiveDependencies: close menu when datasource selection changes
  useEffect(() => {
    setShowDatasourceMenu(false)
  }, [selectedDatasourceId])

  useEffect(() => {
    const ds = metricsDatasources.find((d) => d.id === selectedDatasourceId)
    if (ds) {
      onDatasourceChanged?.({ id: ds.id, name: ds.name, type: ds.type })
    }
  }, [metricsDatasources, onDatasourceChanged, selectedDatasourceId])

  useEffect(() => {
    function handleDocumentClick(event: MouseEvent) {
      const target = event.target as Node
      if (!datasourceMenuRef.current?.contains(target)) {
        setShowDatasourceMenu(false)
      }
    }

    document.addEventListener('click', handleDocumentClick)
    return () => document.removeEventListener('click', handleDocumentClick)
  }, [])

  useEffect(() => {
    return onRefresh(() => {
      if (query.trim() && selectedDatasourceId && result?.status === 'success') {
        void runQuery()
      }
    })
  }, [onRefresh, query, result?.status, runQuery, selectedDatasourceId])

  function handleKeydown(event: React.KeyboardEvent<HTMLDivElement>) {
    if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
      event.preventDefault()
      void runQuery()
    }
  }

  function selectDatasource(datasourceId: string) {
    const prevDs = metricsDatasources.find((d) => d.id === selectedDatasourceId)
    setSelectedDatasourceId(datasourceId)
    setShowDatasourceMenu(false)

    const newDs = metricsDatasources.find((d) => d.id === datasourceId)
    if (newDs) {
      const defaultQuery = getDefaultMetricsQuery(newDs.type)
      if (defaultQuery && (!query.trim() || (prevDs && prevDs.type !== newDs.type))) {
        setQuery(defaultQuery)
      }
    }
  }

  function selectHistoryQuery(q: string) {
    setQuery(q)
    setShowHistory(false)
  }

  function clearHistory() {
    setQueryHistory([])
    sessionStorage.removeItem(METRICS_HISTORY_KEY)
  }

  function toggleDatasourceMenu() {
    if (!hasMetricsDatasources) return
    setShowDatasourceMenu((prev) => !prev)
  }

  function handleUseSuggestion() {
    if (!metricSuggestion) return
    setQuery(applyMetricSuggestion(query, metricSuggestion))
    setError(null)
    setMetricSuggestion(null)
  }

  const favoriteId = `explore::metrics::${query}`
  const queryPlaceholder =
    activeDatasource?.type === 'clickhouse'
      ? 'Enter SQL query...'
      : activeDatasource?.type === 'cloudwatch'
        ? 'Enter CloudWatch query JSON...'
        : activeDatasource?.type === 'elasticsearch'
          ? 'Enter Elasticsearch query JSON...'
          : 'Enter PromQL query...'

  return (
    // biome-ignore lint/a11y/noStaticElementInteractions: panel-level Ctrl/Cmd+Enter to run query
    <section className="flex flex-1 flex-col gap-3" onKeyDown={handleKeydown}>
      <div className="flex justify-end">
        <TimeRangePicker stacked />
      </div>

      <div
        className="flex flex-col gap-2 rounded-md border border-[var(--color-surface-container-high)] bg-[var(--color-surface-container-low)] p-4"
        data-testid="explore-query-editor"
      >
        <div className="flex items-start justify-between text-[11px] leading-normal text-[var(--color-on-surface-variant)]">
          <span className="font-mono font-medium" data-testid="explore-query-language">
            {queryLanguageLabel}
          </span>
          <div ref={datasourceMenuRef} className="relative">
            <button
              id="explore-datasource-btn"
              type="button"
              className="flex cursor-pointer items-center gap-1 border-none bg-transparent p-0 text-[11px] font-normal text-[var(--color-on-surface-variant)] transition-colors duration-[var(--motion-fast)] ease-[var(--ease-standard)] disabled:cursor-not-allowed"
              data-testid="explore-datasource-btn"
              disabled={!hasMetricsDatasources}
              onClick={toggleDatasourceMenu}
              title={
                activeDatasource ? activeDatasource.name : 'No metrics datasource configured'
              }
            >
              <span data-testid="explore-query-datasource">
                {activeDatasource?.name ?? 'No metrics datasource'}
              </span>
              {showDatasourceMenu ? <ChevronUp size={12} /> : <ChevronDown size={12} />}
            </button>

            {showDatasourceMenu && hasMetricsDatasources ? (
              <div
                className="absolute top-full right-0 z-[110] mt-1.5 max-h-[280px] min-w-[240px] overflow-y-auto rounded-md shadow-[var(--shadow-md)]"
                style={{
                  backgroundColor: 'var(--color-surface-bright)',
                  border: '1px solid var(--color-stroke-subtle)',
                }}
              >
                {metricsDatasources.map((ds: DataSource) => (
                  <button
                    key={ds.id}
                    type="button"
                    className="flex w-full cursor-pointer items-center gap-2.5 border-none bg-transparent px-3 py-2.5 text-left transition"
                    style={{
                      color: 'var(--color-on-surface)',
                      backgroundColor:
                        ds.id === selectedDatasourceId ? 'var(--selected-fill)' : 'transparent',
                    }}
                    onClick={() => selectDatasource(ds.id)}
                    onMouseEnter={(event) => {
                      event.currentTarget.style.backgroundColor =
                        'var(--color-surface-container-high)'
                    }}
                    onMouseLeave={(event) => {
                      event.currentTarget.style.backgroundColor =
                        ds.id === selectedDatasourceId ? 'var(--selected-fill)' : 'transparent'
                    }}
                  >
                    <img
                      src={getTypeLogo(ds.type)}
                      alt={`${dataSourceTypeLabels[ds.type]} logo`}
                      className="h-[18px] w-[18px] shrink-0 object-contain"
                    />
                    <div className="flex min-w-0 flex-col gap-px">
                      <strong className="text-sm font-semibold text-[var(--color-on-surface)]">
                        {ds.name}
                      </strong>
                      <span className="text-xs text-[var(--color-outline)]">
                        {dataSourceTypeLabels[ds.type]}
                      </span>
                    </div>
                    {ds.id === selectedDatasourceId ? (
                      <Check size={14} className="ml-auto text-[var(--color-primary)]" />
                    ) : null}
                  </button>
                ))}
              </div>
            ) : null}
          </div>
        </div>

        <MonacoQueryEditor
          value={query}
          onChange={setQuery}
          onSubmit={() => void runQuery()}
          disabled={!hasMetricsDatasources}
          compact={compactQueryEditor}
          height={compactQueryEditor ? 40 : 160}
          placeholder={queryPlaceholder}
        />

        {error ? (
          <div
            className="flex items-start justify-between gap-2 rounded-md p-3"
            style={{
              backgroundColor:
                'color-mix(in srgb, var(--color-error) 22%, var(--color-surface-container-low))',
            }}
            data-testid="explore-query-error"
          >
            <p className="m-0 text-[13px] text-[var(--color-on-surface)]">{error}</p>
            {metricSuggestion ? (
              <Button
                size="sm"
                className="h-auto shrink-0 rounded-md px-2 py-2 text-xs font-semibold"
                data-testid="explore-query-suggestion-btn"
                onClick={handleUseSuggestion}
              >
                Use suggestion
              </Button>
            ) : null}
          </div>
        ) : null}

        <div className="flex flex-wrap items-center gap-2">
          <Button
            size="sm"
            className={`h-auto rounded-md px-2 py-2 text-xs font-semibold ${
              loading
                ? 'border-none bg-[var(--color-surface-container-high)] text-[var(--color-on-surface-variant)] hover:bg-[var(--color-surface-container-high)] disabled:opacity-100'
                : ''
            }`}
            data-testid="explore-run-query-btn"
            disabled={loading || !query.trim() || !selectedDatasourceId || !hasMetricsDatasources}
            onClick={() => void runQuery()}
          >
            {loading ? 'Running' : 'Run query'}
          </Button>

          <Button
            variant="outline"
            size="sm"
            className={`h-auto rounded-md px-2 py-2 text-xs font-medium ${
              loading
                ? 'border-transparent bg-[var(--color-surface-container-high)] text-[var(--color-on-surface-variant)]'
                : 'bg-[var(--color-surface-container-high)] text-[var(--color-on-surface)]'
            }`}
            data-testid="explore-export-btn"
            disabled={loading || !query.trim() || !selectedDatasourceId}
            onClick={() => setShowExportModal(true)}
          >
            Add to dashboard
          </Button>

          {query.trim() ? (
            <button
              type="button"
              className="inline-flex cursor-pointer items-center gap-1.5 rounded-md border px-2 py-2 text-xs transition"
              style={{
                backgroundColor: isFavorite(favoriteId)
                  ? 'var(--color-primary-muted)'
                  : 'var(--color-surface-container-high)',
                borderColor: isFavorite(favoriteId)
                  ? 'var(--color-primary)'
                  : 'var(--color-stroke-subtle)',
                color: isFavorite(favoriteId)
                  ? 'var(--color-primary)'
                  : 'var(--color-on-surface-variant)',
              }}
              title={isFavorite(favoriteId) ? 'Remove from favorites' : 'Save to favorites'}
              onClick={() =>
                toggleFavorite({
                  id: favoriteId,
                  title: query.length > 40 ? `${query.slice(0, 40)}...` : query,
                  type: 'explore',
                })
              }
            >
              <Star size={14} fill={isFavorite(favoriteId) ? 'currentColor' : 'none'} />
            </button>
          ) : null}

          {queryHistory.length > 0 ? (
            <div className="relative">
              <button
                type="button"
                data-testid="explore-history-btn"
                className="flex cursor-pointer items-center gap-1 border-none bg-transparent text-[11px] transition"
                style={{
                  color: showHistory ? 'var(--color-on-surface)' : 'var(--color-outline)',
                }}
                onClick={() => setShowHistory((prev) => !prev)}
                title="Query history"
              >
                <History size={14} />
                <span>History</span>
              </button>

              {showHistory ? (
                <div
                  className="absolute top-full left-0 z-10 mt-1 max-h-[300px] w-80 overflow-y-auto rounded-md shadow-[var(--shadow-md)] max-md:w-full"
                  style={{
                    backgroundColor: 'var(--color-surface-bright)',
                    border: '1px solid var(--color-stroke-subtle)',
                  }}
                >
                  <div
                    className="flex items-center justify-between px-4 py-3 text-xs font-semibold tracking-wide uppercase"
                    style={{
                      color: 'var(--color-outline)',
                      borderBottom: '1px solid var(--color-stroke-subtle)',
                    }}
                  >
                    <span>Recent Queries</span>
                    <button
                      type="button"
                      className="flex h-6 w-6 cursor-pointer items-center justify-center rounded border-none bg-transparent transition"
                      style={{ color: 'var(--color-outline)' }}
                      onClick={clearHistory}
                      title="Clear history"
                    >
                      <X size={14} />
                    </button>
                  </div>
                  {queryHistory.map((historyQuery, index) => (
                    <button
                      // biome-ignore lint/suspicious/noArrayIndexKey: duplicate queries in history are rare; index disambiguates
                      key={`${historyQuery}-${index}`}
                      type="button"
                      className="block w-full cursor-pointer border-none bg-transparent px-4 py-2.5 text-left transition"
                      style={{ borderBottom: '1px solid var(--color-stroke-subtle)' }}
                      onClick={() => selectHistoryQuery(historyQuery)}
                      onMouseEnter={(event) => {
                        event.currentTarget.style.backgroundColor =
                          'var(--color-surface-container-high)'
                      }}
                      onMouseLeave={(event) => {
                        event.currentTarget.style.backgroundColor = 'transparent'
                      }}
                    >
                      <code className="block truncate font-mono text-xs text-[var(--color-on-surface-variant)]">
                        {historyQuery}
                      </code>
                    </button>
                  ))}
                </div>
              ) : null}
            </div>
          ) : null}

          {lastRanAt !== null && !loading && !error ? (
            <LastRanStatus lastRanAt={lastRanAt} seriesCount={seriesCount} />
          ) : null}
        </div>
      </div>

      <div
        className="relative flex min-h-[400px] flex-1 flex-col overflow-hidden rounded-md border border-[var(--color-surface-container-high)] bg-[var(--color-surface-container-low)]"
        data-testid="explore-query-result"
      >
        {loading ? (
          <QueryLoadingSkeleton />
        ) : hasResults ? (
          <div className="flex flex-1 flex-col gap-2 p-4">
            <div className="flex items-start justify-between text-[13px]">
              <span className="font-medium text-[var(--color-on-surface)]">
                {seriesCount === 1 ? '1 series' : `${seriesCount} series`}
              </span>
              {seriesCount > 30 ? (
                <span className="ml-2 text-xs text-[var(--color-tertiary)]">
                  Tip: Add label filters or use an aggregation like{' '}
                  <code
                    className="rounded px-1"
                    style={{ backgroundColor: 'var(--color-surface)' }}
                  >
                    rate()
                  </code>{' '}
                  or{' '}
                  <code
                    className="rounded px-1"
                    style={{ backgroundColor: 'var(--color-surface)' }}
                  >
                    sum by()
                  </code>{' '}
                  to reduce series count.
                </span>
              ) : null}
            </div>
            <div className="min-h-[280px] flex-1">
              <LineChart series={chartSeries} height={400} group="explore-metrics" />
            </div>
          </div>
        ) : error || (result?.status === 'success' && chartSeries.length === 0) ? (
          <QueryEmptyResult />
        ) : !hasMetricsDatasources ? (
          <div className="flex flex-1 flex-col p-6">
            <p className="m-0 text-[13px] text-[var(--color-on-surface-variant)]">
              No metrics datasource configured.
            </p>
          </div>
        ) : (
          <div className="flex flex-1 flex-col p-6">
            <p className="m-0 text-[13px] text-[var(--color-on-surface-variant)]">
              Write a query and click &quot;Run query&quot; to visualize your metrics.
            </p>
          </div>
        )}
      </div>

      {showExportModal ? (
        <ExportToDashboardModal
          query={query}
          signal="metrics"
          datasourceId={selectedDatasourceId}
          onClose={() => setShowExportModal(false)}
        />
      ) : null}
    </section>
  )
}
