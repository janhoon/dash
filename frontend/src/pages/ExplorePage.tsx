import { useCallback, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router'
import { LogsExplorePanel } from '@/components/LogsExplorePanel'
import { MetricsExplorePanel } from '@/components/MetricsExplorePanel'
import { TracesExplorePanel } from '@/components/TracesExplorePanel'
import { useRegisterCommandContext } from '@/hooks/useRegisterCommandContext'

type ExploreType = 'metrics' | 'logs' | 'traces'

type ExploreDatasourceContext = {
  id: string
  name: string
  type: string
}

const tabs: { key: ExploreType; label: string }[] = [
  { key: 'metrics', label: 'Metrics' },
  { key: 'logs', label: 'Logs' },
  { key: 'traces', label: 'Traces' },
]

function normalizeExploreType(type: string | undefined): ExploreType {
  if (type === 'logs' || type === 'traces') return type
  return 'metrics'
}

export function ExplorePage() {
  const navigate = useNavigate()
  const { type } = useParams<{ type: string }>()
  const activeType = useMemo(() => normalizeExploreType(type), [type])
  const [datasource, setDatasource] = useState<ExploreDatasourceContext | null>(null)

  useRegisterCommandContext({
    viewName: `Explore · ${activeType[0]!.toUpperCase()}${activeType.slice(1)}`,
    viewRoute: `/app/explore/${activeType}`,
    description: `Explore ${activeType} from connected datasources.`,
    ...(datasource
      ? {
          datasourceId: datasource.id,
          datasourceType: datasource.type,
          datasourceName: datasource.name,
        }
      : {}),
  })

  const navigateToTab = useCallback(
    (nextType: ExploreType) => {
      if (nextType === activeType) return
      setDatasource(null)
      navigate(`/app/explore/${nextType}`)
    },
    [activeType, navigate],
  )

  const handleDatasourceChanged = useCallback((payload: ExploreDatasourceContext) => {
    setDatasource((current) => {
      if (
        current &&
        current.id === payload.id &&
        current.name === payload.name &&
        current.type === payload.type
      ) {
        return current
      }
      return payload
    })
  }, [])

  const activeTab = tabs.find((tab) => tab.key === activeType) ?? tabs[0]!
  const subtitle = datasource ? `${activeTab.label} · ${datasource.name}` : activeTab.label

  return (
    <div className="flex min-w-0 flex-1 flex-col gap-3 p-6">
      <header className="flex items-center justify-between">
        <div className="flex flex-col gap-1">
          <h1 className="m-0 font-display text-2xl font-semibold text-[var(--color-on-surface)]">
            Explore
          </h1>
          <p
            className="m-0 text-[13px] text-[var(--color-on-surface-variant)]"
            data-testid="explore-subtitle"
          >
            {subtitle}
          </p>
        </div>
      </header>

      <nav className="flex items-start gap-4" data-testid="explore-tab-nav">
        {tabs.map((tab) => {
          const isActive = activeType === tab.key
          return (
            <button
              key={tab.key}
              type="button"
              className="flex cursor-pointer flex-col gap-1 border-none bg-transparent p-0 text-[13px] transition-colors duration-[var(--motion-fast)] ease-[var(--ease-standard)]"
              style={{
                color: isActive ? 'var(--color-on-surface)' : 'var(--color-on-surface-variant)',
                fontWeight: isActive ? 500 : 400,
              }}
              data-testid={`explore-tab-${tab.key}`}
              onClick={() => navigateToTab(tab.key)}
            >
              <span>{tab.label}</span>
              <span
                className="h-0.5 w-full rounded-[1px]"
                style={{
                  backgroundColor: isActive ? 'var(--color-primary)' : 'transparent',
                }}
              />
            </button>
          )
        })}
      </nav>

      {activeType === 'metrics' ? (
        <MetricsExplorePanel key="metrics" onDatasourceChanged={handleDatasourceChanged} />
      ) : activeType === 'logs' ? (
        <LogsExplorePanel key="logs" onDatasourceChanged={handleDatasourceChanged} />
      ) : (
        <TracesExplorePanel key="traces" onDatasourceChanged={handleDatasourceChanged} />
      )}
    </div>
  )
}
