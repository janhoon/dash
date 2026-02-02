import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref, computed } from 'vue'
import Panel from './Panel.vue'

// Mock state that can be controlled per test
let mockLoading = false
let mockError: string | null = null
let mockChartSeries: any[] = []

// Mock the composables
vi.mock('../composables/useTimeRange', () => ({
  useTimeRange: () => ({
    timeRange: computed(() => ({ start: Date.now() - 3600000, end: Date.now() })),
    onRefresh: vi.fn(() => () => {}),
  }),
}))

vi.mock('../composables/useProm', () => ({
  useProm: () => ({
    chartData: computed(() => ({ series: mockChartSeries })),
    loading: computed(() => mockLoading),
    error: computed(() => mockError),
    fetch: vi.fn(),
  }),
}))

// Mock LineChart component
vi.mock('./LineChart.vue', () => ({
  default: {
    name: 'LineChart',
    props: ['series'],
    template: '<div class="mock-line-chart">LineChart Mock</div>',
  },
}))

describe('Panel', () => {
  const mockPanel = {
    id: '1',
    dashboard_id: 'dashboard-1',
    title: 'Test Panel',
    type: 'line_chart',
    grid_pos: { x: 0, y: 0, w: 6, h: 4 },
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  }

  beforeEach(() => {
    // Reset mock state before each test
    mockLoading = false
    mockError = null
    mockChartSeries = []
    vi.clearAllMocks()
  })

  it('renders panel title', () => {
    const wrapper = mount(Panel, {
      props: { panel: mockPanel },
    })
    expect(wrapper.find('.panel-title').text()).toBe('Test Panel')
  })

  it('displays placeholder when no query configured', () => {
    const wrapper = mount(Panel, {
      props: { panel: mockPanel },
    })
    expect(wrapper.find('.panel-state').exists()).toBe(true)
    expect(wrapper.text()).toContain('No query configured')
  })

  it('emits edit event when edit button is clicked', async () => {
    const wrapper = mount(Panel, {
      props: { panel: mockPanel },
    })
    // Find button by title attribute since we use icons now
    const editBtn = wrapper.findAll('button').find((b) => b.attributes('title') === 'Edit')
    expect(editBtn).toBeDefined()
    await editBtn!.trigger('click')
    expect(wrapper.emitted('edit')).toBeTruthy()
    expect(wrapper.emitted('edit')![0]).toEqual([mockPanel])
  })

  it('emits delete event when delete button is clicked', async () => {
    const wrapper = mount(Panel, {
      props: { panel: mockPanel },
    })
    // Find button by title attribute since we use icons now
    const deleteBtn = wrapper.findAll('button').find((b) => b.attributes('title') === 'Delete')
    expect(deleteBtn).toBeDefined()
    await deleteBtn!.trigger('click')
    expect(wrapper.emitted('delete')).toBeTruthy()
    expect(wrapper.emitted('delete')![0]).toEqual([mockPanel])
  })

  it('shows loading state when fetching data', async () => {
    mockLoading = true

    const panelWithQuery = {
      ...mockPanel,
      query: { promql: 'up' },
    }

    const wrapper = mount(Panel, {
      props: { panel: panelWithQuery },
    })

    expect(wrapper.find('.panel-state').exists()).toBe(true)
    expect(wrapper.text()).toContain('Loading')
  })

  it('shows error state when fetch fails', async () => {
    mockError = 'Query failed'

    const panelWithQuery = {
      ...mockPanel,
      query: { promql: 'up' },
    }

    const wrapper = mount(Panel, {
      props: { panel: panelWithQuery },
    })

    expect(wrapper.find('.panel-error').exists()).toBe(true)
    expect(wrapper.text()).toContain('Query failed')
  })

  it('renders LineChart when data is available', async () => {
    mockChartSeries = [
      {
        name: 'up',
        data: [{ timestamp: 1704067200, value: 1 }],
        labels: { __name__: 'up' },
      },
    ]

    const panelWithQuery = {
      ...mockPanel,
      query: { promql: 'up' },
    }

    const wrapper = mount(Panel, {
      props: { panel: panelWithQuery },
    })

    expect(wrapper.find('.chart-container').exists()).toBe(true)
    expect(wrapper.find('.mock-line-chart').exists()).toBe(true)
  })

  it('shows no data message when query returns empty results', async () => {
    mockChartSeries = []

    const panelWithQuery = {
      ...mockPanel,
      query: { promql: 'nonexistent_metric' },
    }

    const wrapper = mount(Panel, {
      props: { panel: panelWithQuery },
    })

    expect(wrapper.find('.panel-no-data').exists()).toBe(true)
    expect(wrapper.text()).toContain('No data')
  })
})
