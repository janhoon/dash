import { type FormEvent, useState } from 'react'
import {
  addMcpPlugin,
  FIGMA_MCP_PLUGINS,
  MCP_PLUGINS_ADD_LABEL,
  MCP_PLUGINS_EMPTY,
  MCP_PLUGINS_RETRY_LABEL,
  MCP_PLUGINS_SUBTITLE,
  MCP_PLUGINS_TITLE,
  type McpPlugin,
  retryMcpPlugin,
  toggleMcpPlugin,
  UNREACHABLE_SERVER_ERROR,
} from '@/components/settings/mcpPlugins'
import { Button } from '@/components/ui/button'

export type McpPluginsPanelProps = {
  initialPlugins?: McpPlugin[]
}

export function McpPluginsPanel({ initialPlugins = FIGMA_MCP_PLUGINS }: McpPluginsPanelProps) {
  const [plugins, setPlugins] = useState<McpPlugin[]>(initialPlugins)
  const [composerOpen, setComposerOpen] = useState(false)
  const [draftName, setDraftName] = useState('')

  function openComposer() {
    setDraftName('')
    setComposerOpen(true)
  }

  function submitComposer(event: FormEvent) {
    event.preventDefault()
    const next = addMcpPlugin(plugins, draftName)
    if (next === plugins) return
    setPlugins(next)
    setComposerOpen(false)
    setDraftName('')
  }

  return (
    <section
      data-testid="mcp-plugins-panel"
      className="flex min-w-0 flex-1 flex-col gap-4 p-6"
      style={{ color: 'var(--color-on-surface)' }}
    >
      <header className="flex w-full items-start justify-between">
        <div className="flex flex-col gap-1">
          <h1 className="m-0 font-display text-2xl font-semibold text-on-surface">
            {MCP_PLUGINS_TITLE}
          </h1>
          <p className="m-0 text-[13px] text-on-surface-variant">{MCP_PLUGINS_SUBTITLE}</p>
        </div>
        <Button
          type="button"
          data-testid="mcp-plugins-add"
          className="h-auto rounded-lg px-2 py-2 text-xs font-semibold"
          onClick={openComposer}
        >
          {MCP_PLUGINS_ADD_LABEL}
        </Button>
      </header>

      {composerOpen ? (
        <form
          data-testid="mcp-plugins-add-form"
          className="flex items-center gap-2"
          onSubmit={submitComposer}
        >
          <input
            data-testid="mcp-plugins-add-name"
            value={draftName}
            onChange={(event) => setDraftName(event.target.value)}
            aria-label="Server name"
            placeholder="Server name"
            className="min-w-0 flex-1 rounded-lg border-0 px-3 py-2 text-sm outline-none"
            style={{
              backgroundColor: 'var(--color-surface-container-high)',
              color: 'var(--color-on-surface)',
            }}
          />
          <Button
            type="submit"
            data-testid="mcp-plugins-add-submit"
            className="h-auto rounded-lg px-2 py-2 text-xs font-semibold"
          >
            {MCP_PLUGINS_ADD_LABEL}
          </Button>
        </form>
      ) : null}

      {plugins.length === 0 ? (
        <p data-testid="mcp-plugins-empty" className="m-0 text-xs text-on-surface-variant">
          {MCP_PLUGINS_EMPTY}.
        </p>
      ) : (
        <ul data-testid="mcp-plugins-list" className="m-0 flex list-none flex-col gap-2 p-0">
          {plugins.map((plugin) => (
            <McpPluginRow
              key={plugin.id}
              plugin={plugin}
              onToggle={() =>
                setPlugins((current) =>
                  current.map((row) => (row.id === plugin.id ? toggleMcpPlugin(row) : row)),
                )
              }
              onRetry={() =>
                setPlugins((current) =>
                  current.map((row) => (row.id === plugin.id ? retryMcpPlugin(row) : row)),
                )
              }
            />
          ))}
        </ul>
      )}
    </section>
  )
}

function McpPluginRow({
  plugin,
  onToggle,
  onRetry,
}: {
  plugin: McpPlugin
  onToggle: () => void
  onRetry: () => void
}) {
  return (
    <li
      data-testid="mcp-plugin-row"
      data-plugin-id={plugin.id}
      className="flex items-center gap-3 rounded-lg p-4"
      style={{
        backgroundColor: 'var(--color-surface-container-low)',
        border: '1px solid var(--color-surface-container-high)',
      }}
    >
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <p className="m-0 text-sm font-medium text-on-surface">{plugin.name}</p>
        <PluginStatus plugin={plugin} onRetry={onRetry} />
      </div>
      <PluginChip plugin={plugin} onToggle={onToggle} />
    </li>
  )
}

function PluginStatus({ plugin, onRetry }: { plugin: McpPlugin; onRetry: () => void }) {
  switch (plugin.state) {
    case 'on':
    case 'off':
      return <p className="m-0 text-xs text-on-surface-variant">{plugin.detail}</p>
    case 'error':
      return (
        <div className="flex flex-wrap items-center gap-2">
          <p className="m-0 text-xs text-error">{UNREACHABLE_SERVER_ERROR}</p>
          <button
            type="button"
            data-testid="mcp-plugin-retry"
            className="cursor-pointer rounded-lg border-0 px-1.5 py-1.5 text-[11px] font-medium"
            style={{
              backgroundColor: 'var(--color-surface-container-high)',
              color: 'var(--color-on-surface-variant)',
            }}
            onClick={onRetry}
          >
            {MCP_PLUGINS_RETRY_LABEL}
          </button>
        </div>
      )
    default: {
      const _exhaustive: never = plugin
      return _exhaustive
    }
  }
}

function PluginChip({ plugin, onToggle }: { plugin: McpPlugin; onToggle: () => void }) {
  switch (plugin.state) {
    case 'on':
    case 'error':
      return (
        <button
          type="button"
          data-testid="mcp-plugin-toggle"
          aria-pressed="true"
          aria-label={`${plugin.name} On`}
          className="cursor-pointer rounded-lg border-0 p-1.5 text-[11px] font-medium"
          style={{ backgroundColor: 'var(--color-primary)', color: '#0B0D0F' }}
          onClick={onToggle}
        >
          On
        </button>
      )
    case 'off':
      return (
        <button
          type="button"
          data-testid="mcp-plugin-toggle"
          aria-pressed="false"
          aria-label={`${plugin.name} Off`}
          className="cursor-pointer rounded-lg border-0 p-1.5 text-[11px] font-medium"
          style={{
            backgroundColor: 'var(--color-surface-container-high)',
            color: 'var(--color-on-surface-variant)',
          }}
          onClick={onToggle}
        >
          Off
        </button>
      )
    default: {
      const _exhaustive: never = plugin
      return _exhaustive
    }
  }
}
