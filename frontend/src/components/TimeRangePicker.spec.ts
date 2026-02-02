import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import TimeRangePicker from './TimeRangePicker.vue'
import { useTimeRange } from '../composables/useTimeRange'

describe('TimeRangePicker', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-02-02T12:00:00Z'))
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.useRealTimers()
    // Clean up shared state
    const { cleanup, setPreset, setRefreshInterval } = useTimeRange()
    cleanup()
    setPreset('1h') // Reset to default
    setRefreshInterval('off')
  })

  it('should render with default time range display', () => {
    const wrapper = mount(TimeRangePicker)

    expect(wrapper.find('.time-display').exists()).toBe(true)
    expect(wrapper.find('.display-text').text()).toBe('Last 1 hour')
  })

  it('should render refresh button', () => {
    const wrapper = mount(TimeRangePicker)

    expect(wrapper.find('.refresh-btn').exists()).toBe(true)
  })

  it('should render refresh interval selector', () => {
    const wrapper = mount(TimeRangePicker)

    const select = wrapper.find('.refresh-interval-selector select')
    expect(select.exists()).toBe(true)

    const options = select.findAll('option')
    expect(options).toHaveLength(5) // Off, 5s, 15s, 30s, 1m
  })

  it('should toggle dropdown when clicking time display', async () => {
    const wrapper = mount(TimeRangePicker)

    expect(wrapper.find('.dropdown').exists()).toBe(false)

    await wrapper.find('.time-display').trigger('click')
    expect(wrapper.find('.dropdown').exists()).toBe(true)

    await wrapper.find('.time-display').trigger('click')
    expect(wrapper.find('.dropdown').exists()).toBe(false)
  })

  it('should display preset options in dropdown', async () => {
    const wrapper = mount(TimeRangePicker)

    await wrapper.find('.time-display').trigger('click')

    const presetItems = wrapper.findAll('.preset-item')
    // 7 presets + 1 custom range button
    expect(presetItems.length).toBeGreaterThanOrEqual(7)

    expect(wrapper.text()).toContain('Last 5 minutes')
    expect(wrapper.text()).toContain('Last 15 minutes')
    expect(wrapper.text()).toContain('Last 30 minutes')
    expect(wrapper.text()).toContain('Last 1 hour')
    expect(wrapper.text()).toContain('Last 6 hours')
    expect(wrapper.text()).toContain('Last 24 hours')
    expect(wrapper.text()).toContain('Last 7 days')
  })

  it('should select preset and close dropdown', async () => {
    const wrapper = mount(TimeRangePicker)

    await wrapper.find('.time-display').trigger('click')

    // Find and click '5 minutes' preset
    const presetButtons = wrapper.findAll('.preset-item')
    const fiveMinButton = presetButtons.find(btn => btn.text() === 'Last 5 minutes')
    expect(fiveMinButton).toBeDefined()

    await fiveMinButton!.trigger('click')

    // Dropdown should close
    expect(wrapper.find('.dropdown').exists()).toBe(false)

    // Display should update
    expect(wrapper.find('.display-text').text()).toBe('Last 5 minutes')
  })

  it('should show custom range form when clicking custom range option', async () => {
    const wrapper = mount(TimeRangePicker)

    await wrapper.find('.time-display').trigger('click')

    const customRangeBtn = wrapper.find('.custom-range-btn')
    expect(customRangeBtn.exists()).toBe(true)

    await customRangeBtn.trigger('click')

    // Should show custom range form
    expect(wrapper.find('.custom-range-form').exists()).toBe(true)
    expect(wrapper.find('#custom-from').exists()).toBe(true)
    expect(wrapper.find('#custom-to').exists()).toBe(true)
  })

  it('should apply custom range', async () => {
    const wrapper = mount(TimeRangePicker)

    await wrapper.find('.time-display').trigger('click')
    await wrapper.find('.custom-range-btn').trigger('click')

    // Set custom dates
    const fromInput = wrapper.find('#custom-from')
    const toInput = wrapper.find('#custom-to')

    await fromInput.setValue('2026-02-01T10:00')
    await toInput.setValue('2026-02-02T14:00')

    // Click apply
    await wrapper.find('.btn-apply').trigger('click')

    // Dropdown should close
    expect(wrapper.find('.dropdown').exists()).toBe(false)

    // Display should show custom range
    expect(wrapper.find('.display-text').text()).toContain('2026-02-01')
    expect(wrapper.find('.display-text').text()).toContain('2026-02-02')
  })

  it('should show error when start time is after end time', async () => {
    const wrapper = mount(TimeRangePicker)

    await wrapper.find('.time-display').trigger('click')
    await wrapper.find('.custom-range-btn').trigger('click')

    // Set invalid dates (start after end)
    await wrapper.find('#custom-from').setValue('2026-02-02T14:00')
    await wrapper.find('#custom-to').setValue('2026-02-01T10:00')

    // Click apply
    await wrapper.find('.btn-apply').trigger('click')

    // Should show error
    expect(wrapper.find('.error-message').exists()).toBe(true)
    expect(wrapper.find('.error-message').text()).toContain('Start time must be before end time')

    // Dropdown should still be open
    expect(wrapper.find('.dropdown').exists()).toBe(true)
  })

  it('should cancel custom range and go back to presets', async () => {
    const wrapper = mount(TimeRangePicker)

    await wrapper.find('.time-display').trigger('click')
    await wrapper.find('.custom-range-btn').trigger('click')

    expect(wrapper.find('.custom-range-form').exists()).toBe(true)

    await wrapper.find('.btn-cancel').trigger('click')

    // Should go back to presets
    expect(wrapper.find('.custom-range-form').exists()).toBe(false)
    expect(wrapper.find('.preset-list').exists()).toBe(true)
  })

  it('should change refresh interval', async () => {
    const wrapper = mount(TimeRangePicker)

    const select = wrapper.find('.refresh-interval-selector select')

    await select.setValue('5s')

    const { refreshIntervalValue } = useTimeRange()
    expect(refreshIntervalValue.value).toBe('5s')
  })

  it('should highlight selected preset', async () => {
    const wrapper = mount(TimeRangePicker)

    // Set to 5m preset first
    const { setPreset } = useTimeRange()
    setPreset('5m')
    await wrapper.vm.$nextTick()

    await wrapper.find('.time-display').trigger('click')

    const selectedItem = wrapper.find('.preset-item.selected')
    expect(selectedItem.exists()).toBe(true)
    expect(selectedItem.text()).toBe('Last 5 minutes')
  })

  it('should call refresh when clicking refresh button', async () => {
    const wrapper = mount(TimeRangePicker)
    const { onRefresh } = useTimeRange()

    const callback = vi.fn()
    onRefresh(callback)

    await wrapper.find('.refresh-btn').trigger('click')

    expect(callback).toHaveBeenCalled()
  })
})
