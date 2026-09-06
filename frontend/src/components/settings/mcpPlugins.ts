export const UNREACHABLE_SERVER_ERROR = "Couldn't reach this server"
export const MCP_PLUGINS_TITLE = 'MCP / Plugins'
export const MCP_PLUGINS_SUBTITLE = 'Tools the agent can call. Quiet list, one gold enable.'
export const MCP_PLUGINS_ADD_LABEL = 'Add server'
export const MCP_PLUGINS_EMPTY = 'No servers yet'
export const MCP_PLUGINS_OFF_DETAIL = 'Off · not configured'
export const MCP_PLUGINS_RETRY_LABEL = 'Retry'

export type McpPlugin =
  | { id: string; name: string; state: 'on'; detail: string }
  | { id: string; name: string; state: 'off'; detail: string }
  | { id: string; name: string; state: 'error'; detail: string }

export const FIGMA_MCP_PLUGINS: McpPlugin[] = [
  {
    id: 'victoria-metrics',
    name: 'Victoria Metrics',
    state: 'on',
    detail: 'Connected · queries + labels',
  },
  {
    id: 'github',
    name: 'GitHub',
    state: 'on',
    detail: 'Connected · issues, PRs',
  },
  {
    id: 'pagerduty',
    name: 'PagerDuty',
    state: 'off',
    detail: MCP_PLUGINS_OFF_DETAIL,
  },
]

export function toggleMcpPlugin(plugin: McpPlugin): McpPlugin {
  switch (plugin.state) {
    case 'on':
      return { id: plugin.id, name: plugin.name, state: 'off', detail: MCP_PLUGINS_OFF_DETAIL }
    case 'off':
    case 'error':
      return {
        id: plugin.id,
        name: plugin.name,
        state: 'on',
        detail: connectedDetail(plugin.detail),
      }
    default: {
      const _exhaustive: never = plugin
      return _exhaustive
    }
  }
}

export function retryMcpPlugin(plugin: McpPlugin): McpPlugin {
  switch (plugin.state) {
    case 'error':
      return { id: plugin.id, name: plugin.name, state: 'on', detail: plugin.detail }
    case 'on':
    case 'off':
      return plugin
    default: {
      const _exhaustive: never = plugin
      return _exhaustive
    }
  }
}

export function addMcpPlugin(plugins: McpPlugin[], name: string): McpPlugin[] {
  const trimmed = name.trim()
  if (!trimmed) return plugins
  return [
    ...plugins,
    {
      id: uniquePluginId(plugins, slugify(trimmed)),
      name: trimmed,
      state: 'off',
      detail: MCP_PLUGINS_OFF_DETAIL,
    },
  ]
}

function connectedDetail(detail: string): string {
  return detail.startsWith('Off') ? 'Connected' : detail
}

function slugify(name: string): string {
  const slug = name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return slug || 'server'
}

function uniquePluginId(plugins: McpPlugin[], base: string): string {
  const taken = new Set(plugins.map((plugin) => plugin.id))
  if (!taken.has(base)) return base
  let n = 2
  while (taken.has(`${base}-${n}`)) n += 1
  return `${base}-${n}`
}
