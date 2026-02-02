import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createWebHistory, type Router } from 'vue-router'
import Explore from './Explore.vue'

// Mock the composables
vi.mock('../composables/useProm', () => ({
  queryPrometheus: vi.fn(),
  transformToChartData: vi.fn(),
  fetchMetrics: vi.fn().mockResolvedValue([]),
  fetchLabels: vi.fn().mockResolvedValue([]),
  fetchLabelValues: vi.fn().mockResolvedValue([])
}))

vi.mock('../composables/useTimeRange', () => ({
  useTimeRange: () => ({
    timeRange: { value: { start: Date.now() - 3600000, end: Date.now() } },
    onRefresh: vi.fn(() => () => {}),
    displayText: { value: 'Last 1 hour' },
    selectedPreset: { value: '1h' },
    isCustomRange: { value: false },
    refreshIntervalValue: { value: 'off' },
    lastRefreshTime: { value: Date.now() },
    isRefreshing: { value: false },
    presets: [],
    refreshIntervals: [],
    setPreset: vi.fn(),
    setCustomRange: vi.fn(),
    setRefreshInterval: vi.fn(),
    refresh: vi.fn(),
  })
}))

// Mock LineChart component
vi.mock('../components/LineChart.vue', () => ({
  default: {
    name: 'LineChart',
    props: ['series', 'height'],
    template: '<div class="mock-line-chart">LineChart Mock</div>'
  }
}))

// Mock TimeRangePicker component
vi.mock('../components/TimeRangePicker.vue', () => ({
  default: {
    name: 'TimeRangePicker',
    template: '<div class="mock-time-range-picker">TimeRangePicker Mock</div>'
  }
}))

import { queryPrometheus, transformToChartData } from '../composables/useProm'

let router: Router

beforeEach(() => {
  router = createRouter({
    history: createWebHistory(),
    routes: [{ path: '/explore', component: Explore }]
  })

  // Clear session storage
  sessionStorage.clear()

  // Reset mocks
  vi.clearAllMocks()
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('Explore', () => {
  it('renders the explore page with header', () => {
    const wrapper = mount(Explore, {
      global: {
        plugins: [router]
      }
    })

    expect(wrapper.find('.explore-header h1').text()).toBe('Explore')
    expect(wrapper.find('.mock-time-range-picker').exists()).toBe(true)
  })

  it('displays query builder component', async () => {
    const wrapper = mount(Explore, {
      global: {
        plugins: [router]
      }
    })
    await flushPromises()

    // QueryBuilder component is now used
    expect(wrapper.findComponent({ name: 'QueryBuilder' }).exists()).toBe(true)
  })

  it('displays Run Query button', () => {
    const wrapper = mount(Explore, {
      global: {
        plugins: [router]
      }
    })

    const runButton = wrapper.find('.btn-run')
    expect(runButton.exists()).toBe(true)
    expect(runButton.text()).toContain('Run Query')
  })

  it('disables Run Query button when query is empty', () => {
    const wrapper = mount(Explore, {
      global: {
        plugins: [router]
      }
    })

    const runButton = wrapper.find('.btn-run')
    expect(runButton.attributes('disabled')).toBeDefined()
  })

  it('enables Run Query button when query is entered', async () => {
    const wrapper = mount(Explore, {
      global: {
        plugins: [router]
      }
    })
    await flushPromises()

    // Simulate QueryBuilder emitting an update
    const queryBuilder = wrapper.findComponent({ name: 'QueryBuilder' })
    await queryBuilder.vm.$emit('update:modelValue', 'up')
    await wrapper.vm.$nextTick()

    const runButton = wrapper.find('.btn-run')
    expect(runButton.attributes('disabled')).toBeUndefined()
  })

  it('shows empty state initially', () => {
    const wrapper = mount(Explore, {
      global: {
        plugins: [router]
      }
    })

    const emptyState = wrapper.find('.empty-state')
    expect(emptyState.exists()).toBe(true)
    expect(emptyState.text()).toContain('Write a PromQL query')
  })

  it('executes query on button click', async () => {
    const mockResult = {
      status: 'success' as const,
      data: {
        resultType: 'matrix',
        result: [
          {
            metric: { __name__: 'up', instance: 'localhost:9090' },
            values: [[1609459200, '1'], [1609459215, '1']]
          }
        ]
      }
    }

    vi.mocked(queryPrometheus).mockResolvedValue(mockResult)
    vi.mocked(transformToChartData).mockReturnValue({
      series: [{
        name: 'up{instance="localhost:9090"}',
        data: [{ timestamp: 1609459200, value: 1 }, { timestamp: 1609459215, value: 1 }],
        labels: { __name__: 'up', instance: 'localhost:9090' }
      }]
    })

    const wrapper = mount(Explore, {
      global: {
        plugins: [router]
      }
    })
    await flushPromises()

    // Simulate QueryBuilder emitting an update
    const queryBuilder = wrapper.findComponent({ name: 'QueryBuilder' })
    await queryBuilder.vm.$emit('update:modelValue', 'up')
    await wrapper.vm.$nextTick()

    const runButton = wrapper.find('.btn-run')
    await runButton.trigger('click')
    await flushPromises()

    expect(queryPrometheus).toHaveBeenCalledWith('up', expect.any(Number), expect.any(Number), expect.any(Number))
  })

  it('shows error when query fails', async () => {
    vi.mocked(queryPrometheus).mockResolvedValue({
      status: 'error',
      error: 'invalid query'
    })

    const wrapper = mount(Explore, {
      global: {
        plugins: [router]
      }
    })
    await flushPromises()

    // Simulate QueryBuilder emitting an update
    const queryBuilder = wrapper.findComponent({ name: 'QueryBuilder' })
    await queryBuilder.vm.$emit('update:modelValue', 'invalid{')
    await wrapper.vm.$nextTick()

    const runButton = wrapper.find('.btn-run')
    await runButton.trigger('click')
    await flushPromises()

    const errorDiv = wrapper.find('.query-error')
    expect(errorDiv.exists()).toBe(true)
    expect(errorDiv.text()).toContain('invalid query')
  })

  it('shows loading state while query executes', async () => {
    let resolveQuery: (value: any) => void
    const queryPromise = new Promise((resolve) => {
      resolveQuery = resolve
    })
    vi.mocked(queryPrometheus).mockReturnValue(queryPromise as any)

    const wrapper = mount(Explore, {
      global: {
        plugins: [router]
      }
    })
    await flushPromises()

    // Simulate QueryBuilder emitting an update
    const queryBuilder = wrapper.findComponent({ name: 'QueryBuilder' })
    await queryBuilder.vm.$emit('update:modelValue', 'up')
    await wrapper.vm.$nextTick()

    const runButton = wrapper.find('.btn-run')
    await runButton.trigger('click')

    // Check loading state
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.loading-state').exists()).toBe(true)
    expect(wrapper.find('.loading-spinner').exists()).toBe(true)

    // Resolve the query
    resolveQuery!({ status: 'success', data: { resultType: 'matrix', result: [] } })
    vi.mocked(transformToChartData).mockReturnValue({ series: [] })
    await flushPromises()

    expect(wrapper.find('.loading-state').exists()).toBe(false)
  })

  it('displays chart when query returns data', async () => {
    const mockResult = {
      status: 'success' as const,
      data: {
        resultType: 'matrix',
        result: [
          {
            metric: { __name__: 'up' },
            values: [[1609459200, '1']]
          }
        ]
      }
    }

    vi.mocked(queryPrometheus).mockResolvedValue(mockResult)
    vi.mocked(transformToChartData).mockReturnValue({
      series: [{
        name: 'up',
        data: [{ timestamp: 1609459200, value: 1 }],
        labels: { __name__: 'up' }
      }]
    })

    const wrapper = mount(Explore, {
      global: {
        plugins: [router]
      }
    })
    await flushPromises()

    // Simulate QueryBuilder emitting an update
    const queryBuilder = wrapper.findComponent({ name: 'QueryBuilder' })
    await queryBuilder.vm.$emit('update:modelValue', 'up')
    await wrapper.vm.$nextTick()

    const runButton = wrapper.find('.btn-run')
    await runButton.trigger('click')
    await flushPromises()

    expect(wrapper.find('.results-container').exists()).toBe(true)
    expect(wrapper.find('.mock-line-chart').exists()).toBe(true)
    expect(wrapper.find('.result-count').text()).toContain('1 series')
  })

  it('shows no data message when query returns empty result', async () => {
    vi.mocked(queryPrometheus).mockResolvedValue({
      status: 'success',
      data: { resultType: 'matrix', result: [] }
    })
    vi.mocked(transformToChartData).mockReturnValue({ series: [] })

    const wrapper = mount(Explore, {
      global: {
        plugins: [router]
      }
    })
    await flushPromises()

    // Simulate QueryBuilder emitting an update
    const queryBuilder = wrapper.findComponent({ name: 'QueryBuilder' })
    await queryBuilder.vm.$emit('update:modelValue', 'nonexistent_metric')
    await wrapper.vm.$nextTick()

    const runButton = wrapper.find('.btn-run')
    await runButton.trigger('click')
    await flushPromises()

    const emptyState = wrapper.find('.empty-state')
    expect(emptyState.exists()).toBe(true)
    expect(emptyState.text()).toContain('No data returned')
  })

  it('saves successful queries to history', async () => {
    const mockResult = {
      status: 'success' as const,
      data: { resultType: 'matrix', result: [] }
    }

    vi.mocked(queryPrometheus).mockResolvedValue(mockResult)
    vi.mocked(transformToChartData).mockReturnValue({ series: [] })

    const wrapper = mount(Explore, {
      global: {
        plugins: [router]
      }
    })
    await flushPromises()

    // Simulate QueryBuilder emitting an update
    const queryBuilder = wrapper.findComponent({ name: 'QueryBuilder' })
    await queryBuilder.vm.$emit('update:modelValue', 'up')
    await wrapper.vm.$nextTick()

    const runButton = wrapper.find('.btn-run')
    await runButton.trigger('click')
    await flushPromises()

    // Check session storage
    const history = JSON.parse(sessionStorage.getItem('explore_query_history') || '[]')
    expect(history).toContain('up')
  })

  it('shows history dropdown when history button is clicked', async () => {
    // Set up some history
    sessionStorage.setItem('explore_query_history', JSON.stringify(['up', 'node_cpu']))

    const wrapper = mount(Explore, {
      global: {
        plugins: [router]
      }
    })

    // Wait for component to load history
    await wrapper.vm.$nextTick()

    const historyBtn = wrapper.find('.history-btn')
    expect(historyBtn.exists()).toBe(true)

    await historyBtn.trigger('click')

    const dropdown = wrapper.find('.history-dropdown')
    expect(dropdown.exists()).toBe(true)
    expect(dropdown.text()).toContain('up')
    expect(dropdown.text()).toContain('node_cpu')
  })

  it('selects query from history', async () => {
    sessionStorage.setItem('explore_query_history', JSON.stringify(['up', 'node_cpu']))

    const wrapper = mount(Explore, {
      global: {
        plugins: [router]
      }
    })
    await flushPromises()

    const historyBtn = wrapper.find('.history-btn')
    await historyBtn.trigger('click')

    const historyItems = wrapper.findAll('.history-item')
    await historyItems[0].trigger('click')

    // The query should be selected - verify the QueryBuilder received the value
    const queryBuilder = wrapper.findComponent({ name: 'QueryBuilder' })
    expect(queryBuilder.props('modelValue')).toBe('up')
  })

  it('clears history when clear button is clicked', async () => {
    sessionStorage.setItem('explore_query_history', JSON.stringify(['up', 'node_cpu']))

    const wrapper = mount(Explore, {
      global: {
        plugins: [router]
      }
    })

    await wrapper.vm.$nextTick()

    const historyBtn = wrapper.find('.history-btn')
    await historyBtn.trigger('click')

    const clearBtn = wrapper.find('.clear-history-btn')
    await clearBtn.trigger('click')

    // History should be cleared
    expect(sessionStorage.getItem('explore_query_history')).toBeNull()
  })
})
