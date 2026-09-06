import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { McpPluginsPanel } from './McpPluginsPanel'
import { type McpPlugin, UNREACHABLE_SERVER_ERROR } from './mcpPlugins'

function rowNamed(name: string) {
  const row = screen.getAllByTestId('mcp-plugin-row').find((el) => within(el).queryByText(name))
  expect(row).toBeTruthy()
  return row as HTMLElement
}

function expectGoldInk(el: HTMLElement) {
  const gold =
    el.style.backgroundColor === 'var(--color-primary)' || el.className.includes('bg-primary')
  const ink = el.style.color === '#0B0D0F' || el.className.includes('text-[#0B0D0F]')
  expect(gold).toBe(true)
  expect(ink).toBe(true)
}

function expectQuietChip(el: HTMLElement) {
  expect(el.style.backgroundColor).toBe('var(--color-surface-container-high)')
  expect(el.style.color).toBe('var(--color-on-surface-variant)')
}

function expectNoQueryToasts() {
  expect(screen.queryByRole('status')).toBeNull()
  expect(screen.queryByTestId('toast')).toBeNull()
  expect(screen.queryByTestId('toast-notification')).toBeNull()
}

describe('McpPluginsPanel', () => {
  it('renders the Figma default list with gold On chips and a quiet Off chip', () => {
    render(<McpPluginsPanel />)

    expect(screen.getByTestId('mcp-plugins-panel')).toBeTruthy()
    expect(screen.getByRole('heading', { name: 'MCP / Plugins' })).toBeTruthy()
    expect(screen.getByText('Tools the agent can call. Quiet list, one gold enable.')).toBeTruthy()

    const add = screen.getByTestId('mcp-plugins-add')
    expect(add.textContent).toContain('Add server')
    expectGoldInk(add)

    expect(screen.getByTestId('mcp-plugins-list')).toBeTruthy()
    expect(screen.getAllByTestId('mcp-plugin-row')).toHaveLength(3)

    const victoria = rowNamed('Victoria Metrics')
    expect(within(victoria).getByText('Connected · queries + labels')).toBeTruthy()
    const victoriaChip = within(victoria).getByTestId('mcp-plugin-toggle')
    expect(victoriaChip.textContent).toBe('On')
    expectGoldInk(victoriaChip)

    const github = rowNamed('GitHub')
    expect(within(github).getByText('Connected · issues, PRs')).toBeTruthy()
    const githubChip = within(github).getByTestId('mcp-plugin-toggle')
    expect(githubChip.textContent).toBe('On')
    expectGoldInk(githubChip)

    const pager = rowNamed('PagerDuty')
    expect(within(pager).getByText('Off · not configured')).toBeTruthy()
    const pagerChip = within(pager).getByTestId('mcp-plugin-toggle')
    expect(pagerChip.textContent).toBe('Off')
    expectQuietChip(pagerChip)
  })

  it('shows an empty state and keeps Add server gold when there are no plugins', () => {
    render(<McpPluginsPanel initialPlugins={[]} />)

    const empty = screen.getByTestId('mcp-plugins-empty')
    expect(empty.textContent?.toLowerCase()).toContain('no servers yet')
    expect(screen.queryByTestId('mcp-plugins-list')).toBeNull()
    expectGoldInk(screen.getByTestId('mcp-plugins-add'))
  })

  it('adds an Off row from the inline Add server field', async () => {
    const user = userEvent.setup()
    render(<McpPluginsPanel initialPlugins={[]} />)

    await user.click(screen.getByTestId('mcp-plugins-add'))
    await user.type(screen.getByTestId('mcp-plugins-add-name'), 'Slack')
    await user.keyboard('{Enter}')

    const slack = rowNamed('Slack')
    expect(within(slack).getByText('Off · not configured')).toBeTruthy()
    expectQuietChip(within(slack).getByTestId('mcp-plugin-toggle'))
  })

  it('toggles PagerDuty On with accent ink and Victoria Metrics Off quiet', async () => {
    const user = userEvent.setup()
    render(<McpPluginsPanel />)

    await user.click(within(rowNamed('PagerDuty')).getByTestId('mcp-plugin-toggle'))
    const pagerOn = within(rowNamed('PagerDuty')).getByTestId('mcp-plugin-toggle')
    expect(pagerOn.textContent).toBe('On')
    expectGoldInk(pagerOn)

    await user.click(within(rowNamed('Victoria Metrics')).getByTestId('mcp-plugin-toggle'))
    const victoriaOff = within(rowNamed('Victoria Metrics')).getByTestId('mcp-plugin-toggle')
    expect(victoriaOff.textContent).toBe('Off')
    expectQuietChip(victoriaOff)
  })

  it('retries an error row inline without a toast', async () => {
    const user = userEvent.setup()
    const initialPlugins: McpPlugin[] = [
      {
        id: 'broken',
        name: 'Broken',
        state: 'error',
        detail: 'timeout',
      },
    ]
    render(<McpPluginsPanel initialPlugins={initialPlugins} />)

    const row = rowNamed('Broken')
    expect(within(row).getByText(UNREACHABLE_SERVER_ERROR)).toBeTruthy()
    expect(within(row).getByTestId('mcp-plugin-retry')).toBeTruthy()
    expectNoQueryToasts()

    await user.click(within(row).getByTestId('mcp-plugin-retry'))

    const recovered = rowNamed('Broken')
    expect(within(recovered).queryByTestId('mcp-plugin-retry')).toBeNull()
    expect(within(recovered).queryByText(UNREACHABLE_SERVER_ERROR)).toBeNull()
    const chip = within(recovered).getByTestId('mcp-plugin-toggle')
    expect(chip.textContent).toBe('On')
    expectGoldInk(chip)
    expectNoQueryToasts()
  })
})
