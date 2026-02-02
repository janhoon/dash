import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import CreateDashboardModal from './CreateDashboardModal.vue'
import * as api from '../api/dashboards'

vi.mock('../api/dashboards')

describe('CreateDashboardModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders form fields', () => {
    const wrapper = mount(CreateDashboardModal)
    expect(wrapper.find('input#title').exists()).toBe(true)
    expect(wrapper.find('textarea#description').exists()).toBe(true)
  })

  it('emits close event when cancel is clicked', async () => {
    const wrapper = mount(CreateDashboardModal)
    await wrapper.findAll('button').find(b => b.text() === 'Cancel')?.trigger('click')
    expect(wrapper.emitted('close')).toBeTruthy()
  })

  it('shows error when title is empty', async () => {
    const wrapper = mount(CreateDashboardModal)
    await wrapper.find('form').trigger('submit')
    expect(wrapper.text()).toContain('Title is required')
  })

  it('calls createDashboard API on submit', async () => {
    vi.mocked(api.createDashboard).mockResolvedValue({
      id: '123',
      title: 'New Dashboard',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z'
    })

    const wrapper = mount(CreateDashboardModal)
    await wrapper.find('input#title').setValue('New Dashboard')
    await wrapper.find('textarea#description').setValue('Description')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(api.createDashboard).toHaveBeenCalledWith({
      title: 'New Dashboard',
      description: 'Description'
    })
    expect(wrapper.emitted('created')).toBeTruthy()
  })

  it('shows error on API failure', async () => {
    vi.mocked(api.createDashboard).mockRejectedValue(new Error('Network error'))

    const wrapper = mount(CreateDashboardModal)
    await wrapper.find('input#title').setValue('New Dashboard')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Failed to create dashboard')
  })
})
