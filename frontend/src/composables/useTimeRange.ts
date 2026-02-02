import { ref, computed, readonly, watch } from 'vue'

export interface TimeRange {
  start: number // Unix timestamp in milliseconds
  end: number // Unix timestamp in milliseconds
}

export interface TimeRangePreset {
  label: string
  value: string
  duration: number // Duration in milliseconds
}

export interface RefreshInterval {
  label: string
  value: string
  interval: number // Interval in milliseconds, 0 means off
}

export const TIME_RANGE_PRESETS: TimeRangePreset[] = [
  { label: 'Last 5 minutes', value: '5m', duration: 5 * 60 * 1000 },
  { label: 'Last 15 minutes', value: '15m', duration: 15 * 60 * 1000 },
  { label: 'Last 30 minutes', value: '30m', duration: 30 * 60 * 1000 },
  { label: 'Last 1 hour', value: '1h', duration: 60 * 60 * 1000 },
  { label: 'Last 6 hours', value: '6h', duration: 6 * 60 * 60 * 1000 },
  { label: 'Last 24 hours', value: '24h', duration: 24 * 60 * 60 * 1000 },
  { label: 'Last 7 days', value: '7d', duration: 7 * 24 * 60 * 60 * 1000 },
]

export const REFRESH_INTERVALS: RefreshInterval[] = [
  { label: 'Off', value: 'off', interval: 0 },
  { label: '5s', value: '5s', interval: 5 * 1000 },
  { label: '15s', value: '15s', interval: 15 * 1000 },
  { label: '30s', value: '30s', interval: 30 * 1000 },
  { label: '1m', value: '1m', interval: 60 * 1000 },
  { label: '5m', value: '5m', interval: 5 * 60 * 1000 },
]

// Shared state across all components using this composable
const selectedPreset = ref<string>('1h')
const customRange = ref<TimeRange | null>(null)
const isCustomRange = ref(false)
const refreshIntervalValue = ref<string>('off')
const lastRefreshTime = ref<number>(Date.now())
const isRefreshing = ref(false)
const isPaused = ref(false)

// Callbacks to be invoked on refresh
const refreshCallbacks = new Set<() => void>()

let refreshIntervalId: number | null = null

function calculateTimeRange(): TimeRange {
  if (isCustomRange.value && customRange.value) {
    return customRange.value
  }

  const preset = TIME_RANGE_PRESETS.find(p => p.value === selectedPreset.value)
  if (!preset) {
    // Default to 1 hour if preset not found
    const now = Date.now()
    return { start: now - 60 * 60 * 1000, end: now }
  }

  const now = Date.now()
  return { start: now - preset.duration, end: now }
}

function startAutoRefresh(intervalMs: number) {
  stopAutoRefresh()
  if (intervalMs > 0 && !isPaused.value) {
    refreshIntervalId = window.setInterval(() => {
      triggerRefresh()
    }, intervalMs)
  }
}

async function triggerRefresh() {
  if (isPaused.value) return

  isRefreshing.value = true
  lastRefreshTime.value = Date.now()

  // Execute all refresh callbacks
  const promises = Array.from(refreshCallbacks).map(callback => {
    try {
      const result = callback()
      return result instanceof Promise ? result : Promise.resolve()
    } catch {
      return Promise.resolve()
    }
  })

  await Promise.all(promises)
  isRefreshing.value = false
}

function stopAutoRefresh() {
  if (refreshIntervalId !== null) {
    window.clearInterval(refreshIntervalId)
    refreshIntervalId = null
  }
}

export function useTimeRange() {
  const timeRange = computed(() => calculateTimeRange())

  const displayText = computed(() => {
    if (isCustomRange.value && customRange.value) {
      const start = new Date(customRange.value.start)
      const end = new Date(customRange.value.end)
      return `${formatDateTime(start)} - ${formatDateTime(end)}`
    }

    const preset = TIME_RANGE_PRESETS.find(p => p.value === selectedPreset.value)
    return preset?.label || 'Last 1 hour'
  })

  const refreshInterval = computed(() => {
    return REFRESH_INTERVALS.find(r => r.value === refreshIntervalValue.value) || REFRESH_INTERVALS[0]
  })

  function setPreset(presetValue: string) {
    const preset = TIME_RANGE_PRESETS.find(p => p.value === presetValue)
    if (preset) {
      selectedPreset.value = presetValue
      isCustomRange.value = false
      customRange.value = null
      lastRefreshTime.value = Date.now()
      refreshCallbacks.forEach(callback => callback())
    }
  }

  function setCustomRange(start: number, end: number) {
    customRange.value = { start, end }
    isCustomRange.value = true
    lastRefreshTime.value = Date.now()
    refreshCallbacks.forEach(callback => callback())
  }

  function setRefreshInterval(intervalValue: string) {
    const interval = REFRESH_INTERVALS.find(r => r.value === intervalValue)
    if (interval) {
      refreshIntervalValue.value = intervalValue
      startAutoRefresh(interval.interval)
    }
  }

  function refresh() {
    triggerRefresh()
  }

  function pauseAutoRefresh() {
    isPaused.value = true
    stopAutoRefresh()
  }

  function resumeAutoRefresh() {
    isPaused.value = false
    const interval = REFRESH_INTERVALS.find(r => r.value === refreshIntervalValue.value)
    if (interval && interval.interval > 0) {
      startAutoRefresh(interval.interval)
    }
  }

  function onRefresh(callback: () => void) {
    refreshCallbacks.add(callback)
    // Return cleanup function
    return () => {
      refreshCallbacks.delete(callback)
    }
  }

  function cleanup() {
    stopAutoRefresh()
    refreshCallbacks.clear()
  }

  return {
    // State (readonly)
    timeRange,
    displayText,
    selectedPreset: readonly(selectedPreset),
    isCustomRange: readonly(isCustomRange),
    customRange: readonly(customRange),
    refreshInterval,
    refreshIntervalValue: readonly(refreshIntervalValue),
    lastRefreshTime: readonly(lastRefreshTime),
    isRefreshing: readonly(isRefreshing),
    isPaused: readonly(isPaused),

    // Constants
    presets: TIME_RANGE_PRESETS,
    refreshIntervals: REFRESH_INTERVALS,

    // Actions
    setPreset,
    setCustomRange,
    setRefreshInterval,
    refresh,
    onRefresh,
    cleanup,
    pauseAutoRefresh,
    resumeAutoRefresh,
  }
}

function formatDateTime(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}
