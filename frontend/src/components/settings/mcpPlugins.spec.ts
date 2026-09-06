import { describe, expect, it } from 'vitest'
import {
  addMcpPlugin,
  FIGMA_MCP_PLUGINS,
  retryMcpPlugin,
  toggleMcpPlugin,
  UNREACHABLE_SERVER_ERROR,
} from './mcpPlugins'

describe('toggleMcpPlugin', () => {
  it('turns On into Off with not-configured detail', () => {
    expect(toggleMcpPlugin(FIGMA_MCP_PLUGINS[0]!)).toEqual({
      id: 'victoria-metrics',
      name: 'Victoria Metrics',
      state: 'off',
      detail: 'Off · not configured',
    })
  })

  it('turns Off into On with a Connected detail', () => {
    expect(toggleMcpPlugin(FIGMA_MCP_PLUGINS[2]!)).toEqual({
      id: 'pagerduty',
      name: 'PagerDuty',
      state: 'on',
      detail: 'Connected',
    })
  })

  it('clears error into On and keeps a non-Off detail', () => {
    const plugin = {
      id: 'broken',
      name: 'Broken',
      state: 'error' as const,
      detail: UNREACHABLE_SERVER_ERROR,
    }
    expect(toggleMcpPlugin(plugin)).toEqual({
      id: 'broken',
      name: 'Broken',
      state: 'on',
      detail: UNREACHABLE_SERVER_ERROR,
    })
  })
})

describe('retryMcpPlugin', () => {
  it('turns error into On and keeps detail', () => {
    const plugin = {
      id: 'broken',
      name: 'Broken',
      state: 'error' as const,
      detail: UNREACHABLE_SERVER_ERROR,
    }
    expect(retryMcpPlugin(plugin)).toEqual({
      id: 'broken',
      name: 'Broken',
      state: 'on',
      detail: UNREACHABLE_SERVER_ERROR,
    })
  })

  it('is identity for On and Off', () => {
    expect(retryMcpPlugin(FIGMA_MCP_PLUGINS[0]!)).toBe(FIGMA_MCP_PLUGINS[0])
    expect(retryMcpPlugin(FIGMA_MCP_PLUGINS[2]!)).toBe(FIGMA_MCP_PLUGINS[2])
  })
})

describe('addMcpPlugin', () => {
  it('appends an Off row from a trimmed name', () => {
    expect(addMcpPlugin([], ' Slack ')).toEqual([
      { id: 'slack', name: 'Slack', state: 'off', detail: 'Off · not configured' },
    ])
  })

  it('ignores a blank name and returns the same array', () => {
    expect(addMcpPlugin(FIGMA_MCP_PLUGINS, '   ')).toBe(FIGMA_MCP_PLUGINS)
    expect(addMcpPlugin(FIGMA_MCP_PLUGINS, '')).toBe(FIGMA_MCP_PLUGINS)
  })

  it('makes a colliding slug unique', () => {
    const next = addMcpPlugin(FIGMA_MCP_PLUGINS, 'GitHub')
    expect(next.at(-1)).toEqual({
      id: 'github-2',
      name: 'GitHub',
      state: 'off',
      detail: 'Off · not configured',
    })
  })
})
