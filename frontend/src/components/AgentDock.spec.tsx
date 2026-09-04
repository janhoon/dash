import { act, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'
import { AgentDock } from '@/components/AgentDock'
import { useAiSidebarStore } from '@/stores/aiSidebarStore'

describe('AgentDock', () => {
  beforeEach(() => {
    localStorage.clear()
    useAiSidebarStore.setState({
      isOpen: false,
      pendingContext: null,
      highlightedPanelId: null,
    })
  })

  it('does not render when closed', () => {
    render(<AgentDock />)
    expect(screen.queryByTestId('agent-dock')).toBeNull()
  })

  it('renders Figma placeholder chrome when open', () => {
    useAiSidebarStore.setState({ isOpen: true })
    render(<AgentDock />)

    const dock = screen.getByTestId('agent-dock')
    expect(dock).toBeTruthy()
    expect(dock.style.width).toBe('380px')
    expect(screen.getByText('Agent')).toBeTruthy()
    expect(screen.getByText('⌘J')).toBeTruthy()
    expect(screen.getByText('You')).toBeTruthy()
    expect(screen.getByText('Why did p99 move after 16:00?')).toBeTruthy()
    expect(screen.getByText('Ace')).toBeTruthy()
    expect(
      screen.getByText(
        'Checkout deploy at 16:12. Latency by route shows /pay jumped 80ms. Error ratio flat. Pin that series — do not make a new board.',
      ),
    ).toBeTruthy()
    const pin = screen.getByRole('button', { name: 'Pin /pay p99' })
    expect(pin.className).toContain('bg-primary')
    expect(pin.className).toContain('text-[#0B0D0F]')
    expect(screen.getByPlaceholderText('Ask about this board…')).toBeTruthy()
    expect(screen.queryByTestId('ai-fab')).toBeNull()
    expect(screen.queryByText('Copilot')).toBeNull()
  })

  it('renders composer only when the thread is empty', () => {
    useAiSidebarStore.setState({ isOpen: true })
    render(<AgentDock thread={[]} />)

    expect(screen.getByTestId('agent-dock')).toBeTruthy()
    expect(screen.getByPlaceholderText('Ask about this board…')).toBeTruthy()
    expect(screen.queryByText('You')).toBeNull()
    expect(screen.queryByText('Ace')).toBeNull()
    expect(screen.queryByText('Why did p99 move after 16:00?')).toBeNull()
    expect(screen.queryByRole('button', { name: 'Pin /pay p99' })).toBeNull()
  })

  it('opens and closes when the store toggles', async () => {
    render(<AgentDock />)
    expect(screen.queryByTestId('agent-dock')).toBeNull()

    act(() => {
      useAiSidebarStore.getState().toggle()
    })
    expect(await screen.findByTestId('agent-dock')).toBeTruthy()

    act(() => {
      useAiSidebarStore.getState().toggle()
    })
    await waitFor(() => {
      expect(screen.queryByTestId('agent-dock')).toBeNull()
    })
  })

  it('closes when the store closes', async () => {
    useAiSidebarStore.setState({ isOpen: true })
    render(<AgentDock />)
    expect(screen.getByTestId('agent-dock')).toBeTruthy()

    act(() => {
      useAiSidebarStore.getState().close()
    })
    await waitFor(() => {
      expect(screen.queryByTestId('agent-dock')).toBeNull()
    })
  })
})
