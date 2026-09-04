import { act, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'
import { AgentDock, PLACEHOLDER_THREAD } from '@/components/AgentDock'
import { ShortcutsOverlay } from '@/components/ShortcutsOverlay'
import { useKeyboardShortcutsStore } from '@/lib/keyboardShortcuts'
import { useAiSidebarStore } from '@/stores/aiSidebarStore'

describe('AgentDock', () => {
  beforeEach(() => {
    localStorage.clear()
    useKeyboardShortcutsStore.getState()._reset()
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

  it('renders empty chrome when open by default', () => {
    useAiSidebarStore.setState({ isOpen: true })
    render(<AgentDock />)

    const dock = screen.getByTestId('agent-dock')
    expect(dock).toBeTruthy()
    expect(dock.style.width).toBe('380px')
    expect(screen.getByText('Agent')).toBeTruthy()
    expect(screen.getByText('⌘J')).toBeTruthy()
    expect(screen.getByTestId('agent-dock-thread').className).toContain('overflow-y-auto')
    expect(screen.getByTestId('agent-dock-thread').className).toContain('flex-1')
    expect(screen.getByTestId('agent-dock-thread').className).toContain('min-h-0')
    expect(screen.queryByText('You')).toBeNull()
    expect(screen.queryByText('Ace')).toBeNull()
    expect(screen.queryByText('Why did p99 move after 16:00?')).toBeNull()
    expect(screen.queryByRole('button', { name: 'Pin /pay p99' })).toBeNull()
    const composer = screen.getByPlaceholderText('Ask about this board…')
    expect(composer).toBeTruthy()
    expect(composer).toHaveProperty('readOnly', true)
    expect(screen.queryByTestId('ai-fab')).toBeNull()
    expect(screen.queryByText('Copilot')).toBeNull()
  })

  it('renders Figma placeholder chrome when given the potato thread', () => {
    useAiSidebarStore.setState({ isOpen: true })
    render(<AgentDock thread={PLACEHOLDER_THREAD} />)

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
    const composer = screen.getByPlaceholderText('Ask about this board…')
    expect(composer).toBeTruthy()
    expect(composer).toHaveProperty('readOnly', true)
    expect(screen.queryByTestId('ai-fab')).toBeNull()
    expect(screen.queryByText('Copilot')).toBeNull()
  })

  it('surfaces pendingContext as a user turn when opened', async () => {
    render(<AgentDock />)

    act(() => {
      useAiSidebarStore.getState().open({
        message:
          'Investigate the currently firing alerts: disk-full. Analyze root causes, correlate with recent deployments or changes, assess severity, and suggest remediation steps.',
      })
    })

    expect(
      await screen.findByText(
        'Investigate the currently firing alerts: disk-full. Analyze root causes, correlate with recent deployments or changes, assess severity, and suggest remediation steps.',
      ),
    ).toBeTruthy()
    expect(screen.getByText('You')).toBeTruthy()
    expect(screen.queryByText('Ace')).toBeNull()
    expect(screen.queryByText('Why did p99 move after 16:00?')).toBeNull()
    await waitFor(() => {
      expect(useAiSidebarStore.getState().pendingContext).toBeNull()
    })
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

  it('closes on Escape while open', async () => {
    useAiSidebarStore.setState({ isOpen: true })
    render(<AgentDock />)
    expect(screen.getByTestId('agent-dock')).toBeTruthy()

    act(() => {
      window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    })
    await waitFor(() => {
      expect(screen.queryByTestId('agent-dock')).toBeNull()
    })
    expect(useAiSidebarStore.getState().isOpen).toBe(false)
  })

  it('does not close on Escape when a higher overlay is open', async () => {
    useAiSidebarStore.setState({ isOpen: true })
    render(<AgentDock overlayOpen />)
    expect(screen.getByTestId('agent-dock')).toBeTruthy()

    act(() => {
      window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    })
    expect(screen.getByTestId('agent-dock')).toBeTruthy()
    expect(useAiSidebarStore.getState().isOpen).toBe(true)
  })

  it('does not close on Escape when shortcuts help is open', async () => {
    useAiSidebarStore.setState({ isOpen: true })
    useKeyboardShortcutsStore.getState().setShowHelp(true)
    render(
      <>
        <AgentDock />
        <ShortcutsOverlay />
      </>,
    )
    expect(screen.getByTestId('agent-dock')).toBeTruthy()
    expect(screen.getByRole('dialog', { name: 'Keyboard shortcuts' })).toBeTruthy()

    act(() => {
      window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    })
    await waitFor(() => {
      expect(screen.queryByRole('dialog', { name: 'Keyboard shortcuts' })).toBeNull()
    })
    expect(screen.getByTestId('agent-dock')).toBeTruthy()
    expect(useAiSidebarStore.getState().isOpen).toBe(true)
  })
})
