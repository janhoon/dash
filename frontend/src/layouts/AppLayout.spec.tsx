import { QueryClientProvider } from '@tanstack/react-query'
import { act, render, screen, waitFor } from '@testing-library/react'
import { createMemoryRouter } from 'react-router'
import { RouterProvider } from 'react-router/dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import * as datasourcesApi from '@/api/datasources'
import * as organizationsApi from '@/api/organizations'
import { AGENT_DOCK_WIDTH_PX } from '@/components/AgentDock'
import { AppLayout } from '@/layouts/AppLayout'
import { useKeyboardShortcutsStore } from '@/lib/keyboardShortcuts'
import { PlaceholderPage } from '@/pages/PlaceholderPage'
import { useAiSidebarStore } from '@/stores/aiSidebarStore'
import { useAuthStore } from '@/stores/authStore'
import { useOrgStore } from '@/stores/orgStore'
import { useSidebarStore } from '@/stores/sidebarStore'
import { createTestQueryClient } from '@/test/renderWithProviders'

vi.mock('@/analytics', () => ({
  identifyUser: vi.fn(),
  resetUserAnalytics: vi.fn(),
  trackEvent: vi.fn(),
  getAnalyticsReady: vi.fn(() => false),
  getAnalyticsConsent: vi.fn(() => 'granted'),
  getAnalyticsDntEnabled: vi.fn(() => false),
  getAnalyticsSessionRecordingEnabled: vi.fn(() => false),
  setAnalyticsConsent: vi.fn(),
  setSessionRecordingEnabled: vi.fn(),
  initializeAnalytics: vi.fn(),
  subscribeAnalytics: (listener: () => void) => {
    void listener
    return () => {}
  },
}))

function renderAppLayout(initialPath = '/app') {
  const queryClient = createTestQueryClient()
  const router = createMemoryRouter(
    [
      {
        element: <AppLayout />,
        children: [
          {
            path: '/app',
            element: <PlaceholderPage title="Command Center" />,
          },
        ],
      },
    ],
    { initialEntries: [initialPath] },
  )

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

describe('AppLayout', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    localStorage.clear()
    useSidebarStore.getState()._reset()
    useKeyboardShortcutsStore.getState()._reset()
    useAiSidebarStore.setState({
      isOpen: false,
      pendingContext: null,
      highlightedPanelId: null,
    })
    useOrgStore.setState({ currentOrgId: 'org-1' })
    useAuthStore.setState({
      user: { id: 'u1', email: 'user@example.com', name: 'User', created_at: '', updated_at: '' },
      userOrganizations: [{ id: 'org-1', name: 'Test Org', slug: 'test', role: 'admin' }],
      loading: false,
      initialized: true,
      isAuthenticated: true,
    })
    vi.spyOn(organizationsApi, 'listOrganizations').mockResolvedValue([
      {
        id: 'org-1',
        name: 'Test Org',
        slug: 'test',
        created_at: '',
        updated_at: '',
      },
    ])
    vi.spyOn(datasourcesApi, 'listDataSources').mockResolvedValue([])
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1440 })
  })

  it('renders placeholder route inside app layout with sidebar', async () => {
    renderAppLayout()
    expect(screen.getByRole('heading', { name: 'Command Center' })).toBeTruthy()
    expect(screen.getByRole('main')).toBeTruthy()
    expect(await screen.findByTestId('sidebar')).toBeTruthy()
  })

  it('fetches datasources when org is set', async () => {
    const listSpy = vi.spyOn(datasourcesApi, 'listDataSources').mockResolvedValue([])
    renderAppLayout()
    await waitFor(() => {
      expect(listSpy).toHaveBeenCalledWith('org-1')
    })
  })

  it('shows narrow viewport overlay below 1280px', async () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 })
    renderAppLayout()
    expect(await screen.findByTestId('narrow-viewport-overlay')).toBeTruthy()
  })

  it('keeps the agent dock closed with no FAB', async () => {
    renderAppLayout()
    expect(await screen.findByTestId('sidebar')).toBeTruthy()
    expect(screen.queryByTestId('agent-dock')).toBeNull()
    expect(screen.queryByTestId('ai-fab')).toBeNull()
    expect(screen.getByRole('main').style.marginRight).toBe('0px')
  })

  it('opens the right dock and pushes the board', async () => {
    useAiSidebarStore.setState({ isOpen: true })
    renderAppLayout()
    const dock = await screen.findByTestId('agent-dock')
    expect(dock.style.width).toBe(`${AGENT_DOCK_WIDTH_PX}px`)
    expect(screen.queryByTestId('ai-fab')).toBeNull()
    expect(screen.getByRole('main').style.marginRight).toBe(`${AGENT_DOCK_WIDTH_PX}px`)
    expect(screen.getByPlaceholderText('Ask about this board…')).toHaveProperty('readOnly', true)
    expect(screen.queryByText('Why did p99 move after 16:00?')).toBeNull()
    expect(screen.queryByText('Ace')).toBeNull()
  })

  it('toggles the agent dock with Cmd+J', async () => {
    renderAppLayout()
    expect(await screen.findByTestId('sidebar')).toBeTruthy()
    expect(screen.queryByTestId('agent-dock')).toBeNull()

    act(() => {
      window.dispatchEvent(new KeyboardEvent('keydown', { key: 'j', metaKey: true, bubbles: true }))
    })
    expect(await screen.findByTestId('agent-dock')).toBeTruthy()
    expect(screen.getByRole('main').style.marginRight).toBe(`${AGENT_DOCK_WIDTH_PX}px`)

    act(() => {
      window.dispatchEvent(new KeyboardEvent('keydown', { key: 'j', metaKey: true, bubbles: true }))
    })
    await waitFor(() => {
      expect(screen.queryByTestId('agent-dock')).toBeNull()
    })
    expect(screen.getByRole('main').style.marginRight).toBe('0px')
  })
})
