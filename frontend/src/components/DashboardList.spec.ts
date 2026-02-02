import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import DashboardList from './DashboardList.vue'
import * as api from '../api/dashboards'

vi.mock('../api/dashboards')

const mockDashboards = [
  {
    id: '123e4567-e89b-12d3-a456-426614174000',
    title: 'Test Dashboard',
    description: 'Test description',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z'
  },
  {
    id: '223e4567-e89b-12d3-a456-426614174001',
    title: 'Another Dashboard',
    description: null,
    created_at: '2024-01-02T00:00:00Z',
    updated_at: '2024-01-02T00:00:00Z'
  }
]

describe('DashboardList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('displays loading state initially', () => {
    vi.mocked(api.listDashboards).mockImplementation(() => new Promise(() => {}))
    const wrapper = mount(DashboardList)
    expect(wrapper.text()).toContain('Loading dashboards...')
  })

  it('displays dashboards after loading', async () => {
    vi.mocked(api.listDashboards).mockResolvedValue(mockDashboards)
    const wrapper = mount(DashboardList)
    await flushPromises()
    expect(wrapper.text()).toContain('Test Dashboard')
    expect(wrapper.text()).toContain('Another Dashboard')
  })

  it('displays empty state when no dashboards', async () => {
    vi.mocked(api.listDashboards).mockResolvedValue([])
    const wrapper = mount(DashboardList)
    await flushPromises()
    expect(wrapper.text()).toContain('No dashboards yet')
  })

  it('displays error state on fetch failure', async () => {
    vi.mocked(api.listDashboards).mockRejectedValue(new Error('Network error'))
    const wrapper = mount(DashboardList)
    await flushPromises()
    expect(wrapper.text()).toContain('Failed to load dashboards')
  })

  it('opens create modal when button is clicked', async () => {
    vi.mocked(api.listDashboards).mockResolvedValue([])
    const wrapper = mount(DashboardList)
    await flushPromises()
    await wrapper.find('.btn-primary').trigger('click')
    expect(wrapper.findComponent({ name: 'CreateDashboardModal' }).exists()).toBe(true)
  })

  it('renders dashboard cards with titles', async () => {
    vi.mocked(api.listDashboards).mockResolvedValue(mockDashboards)
    const wrapper = mount(DashboardList)
    await flushPromises()
    const cards = wrapper.findAll('.dashboard-card')
    expect(cards.length).toBe(2)
  })
})
