import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { nextTick } from 'vue'
import DashboardDetailView from './DashboardDetailView.vue'

// Mock vue-router
const mockRouteParams = { id: 'test-dashboard-id' }
const mockPush = vi.fn()

vi.mock('vue-router', () => ({
  useRoute: () => ({
    params: mockRouteParams
  }),
  useRouter: () => ({
    push: mockPush
  })
}))

// Mock api functions
const mockDashboard = {
  id: 'test-dashboard-id',
  title: 'Test Dashboard',
  description: 'Test Description',
  created_at: '2026-02-02T00:00:00Z',
  updated_at: '2026-02-02T00:00:00Z'
}

const mockPanels = [
  {
    id: 'panel-1',
    dashboard_id: 'test-dashboard-id',
    title: 'Panel 1',
    type: 'line_chart',
    grid_pos: { x: 0, y: 0, w: 6, h: 3 },
    query: { promql: 'up' },
    created_at: '2026-02-02T00:00:00Z',
    updated_at: '2026-02-02T00:00:00Z'
  },
  {
    id: 'panel-2',
    dashboard_id: 'test-dashboard-id',
    title: 'Panel 2',
    type: 'line_chart',
    grid_pos: { x: 6, y: 0, w: 6, h: 3 },
    query: { promql: 'process_cpu_seconds_total' },
    created_at: '2026-02-02T00:00:00Z',
    updated_at: '2026-02-02T00:00:00Z'
  }
]

vi.mock('../api/dashboards', () => ({
  getDashboard: vi.fn(() => Promise.resolve(mockDashboard))
}))

vi.mock('../api/panels', () => ({
  listPanels: vi.fn(() => Promise.resolve(mockPanels)),
  deletePanel: vi.fn(() => Promise.resolve()),
  updatePanel: vi.fn(() => Promise.resolve())
}))

// Mock vue-grid-layout
vi.mock('vue-grid-layout', () => ({
  GridLayout: {
    name: 'GridLayout',
    template: '<div class="grid-layout"><slot /></div>',
    props: ['layout', 'colNum', 'rowHeight', 'isDraggable', 'isResizable', 'verticalCompact', 'useCssTransforms', 'responsive', 'breakpoints', 'cols']
  },
  GridItem: {
    name: 'GridItem',
    template: '<div class="grid-item"><slot /></div>',
    props: ['i', 'x', 'y', 'w', 'h', 'minW', 'minH', 'dragAllowFrom', 'dragIgnoreFrom']
  }
}))

// Mock useTimeRange
import { ref, computed } from 'vue'

vi.mock('../composables/useTimeRange', () => ({
  useTimeRange: () => ({
    timeRange: computed(() => ({ start: Date.now() - 3600000, end: Date.now() })),
    displayText: computed(() => 'Last 1 hour'),
    selectedPreset: ref('1h'),
    isCustomRange: ref(false),
    customRange: ref(null),
    refreshInterval: computed(() => ({ label: 'Off', value: 'off', interval: 0 })),
    refreshIntervalValue: ref('off'),
    lastRefreshTime: ref(Date.now()),
    isRefreshing: ref(false),
    isPaused: ref(false),
    presets: [],
    refreshIntervals: [],
    onRefresh: vi.fn(() => () => {}),
    cleanup: vi.fn(),
    pauseAutoRefresh: vi.fn(),
    resumeAutoRefresh: vi.fn(),
    setPreset: vi.fn(),
    setCustomRange: vi.fn(),
    setRefreshInterval: vi.fn(),
    refresh: vi.fn()
  })
}))

// Mock useProm (used by Panel component)
vi.mock('../composables/useProm', () => ({
  useProm: () => ({
    chartData: { value: { series: [] } },
    loading: { value: false },
    error: { value: null },
    fetch: vi.fn()
  })
}))

describe('DashboardDetailView', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-02-02T12:00:00Z'))
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.useRealTimers()
  })

  it('should render dashboard title after loading', async () => {
    const wrapper = mount(DashboardDetailView)
    await flushPromises()

    expect(wrapper.find('h1').text()).toBe('Test Dashboard')
  })

  it('should render panels using grid layout', async () => {
    const wrapper = mount(DashboardDetailView)
    await flushPromises()

    // Should render Panel components (which are inside grid items)
    const panels = wrapper.findAllComponents({ name: 'Panel' })
    expect(panels).toHaveLength(2)
  })

  it('should display loading state initially', () => {
    const wrapper = mount(DashboardDetailView)

    expect(wrapper.find('.loading').exists()).toBe(true)
    expect(wrapper.text()).toContain('Loading dashboard...')
  })

  it('should show back button that navigates to dashboards', async () => {
    const wrapper = mount(DashboardDetailView)
    await flushPromises()

    const backBtn = wrapper.find('.btn-back')
    expect(backBtn.exists()).toBe(true)

    await backBtn.trigger('click')
    expect(mockPush).toHaveBeenCalledWith('/dashboards')
  })

  it('should show Add Panel button', async () => {
    const wrapper = mount(DashboardDetailView)
    await flushPromises()

    const addBtn = wrapper.find('.btn-primary')
    expect(addBtn.exists()).toBe(true)
    expect(addBtn.text()).toContain('Add Panel')
  })

  it('should open panel modal when Add Panel is clicked', async () => {
    const wrapper = mount(DashboardDetailView)
    await flushPromises()

    const addBtn = wrapper.find('.btn-primary')
    await addBtn.trigger('click')

    expect(wrapper.findComponent({ name: 'PanelEditModal' }).exists()).toBe(true)
  })

  it('should render TimeRangePicker', async () => {
    const wrapper = mount(DashboardDetailView)
    await flushPromises()

    expect(wrapper.findComponent({ name: 'TimeRangePicker' }).exists()).toBe(true)
  })

  it('should show empty state when no panels', async () => {
    const { listPanels } = await import('../api/panels')
    vi.mocked(listPanels).mockResolvedValueOnce([])

    const wrapper = mount(DashboardDetailView)
    await flushPromises()

    expect(wrapper.find('.empty').exists()).toBe(true)
    expect(wrapper.text()).toContain('No panels yet')
  })
})
