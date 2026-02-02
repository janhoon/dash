import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useTimeRange, TIME_RANGE_PRESETS, REFRESH_INTERVALS } from './useTimeRange'

describe('useTimeRange', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-02-02T12:00:00Z'))
  })

  afterEach(() => {
    // Clean up between tests - must be done before restoring timers
    const { cleanup, setPreset, setRefreshInterval, resumeAutoRefresh } = useTimeRange()
    cleanup()
    // Reset to default values
    setPreset('1h')
    setRefreshInterval('off')
    resumeAutoRefresh() // Reset paused state
    vi.restoreAllMocks()
    vi.useRealTimers()
  })

  describe('TIME_RANGE_PRESETS', () => {
    it('should have all expected presets', () => {
      expect(TIME_RANGE_PRESETS).toHaveLength(7)
      expect(TIME_RANGE_PRESETS.map(p => p.value)).toEqual([
        '5m', '15m', '30m', '1h', '6h', '24h', '7d'
      ])
    })

    it('should have correct durations', () => {
      const presetMap = Object.fromEntries(TIME_RANGE_PRESETS.map(p => [p.value, p.duration]))
      expect(presetMap['5m']).toBe(5 * 60 * 1000)
      expect(presetMap['15m']).toBe(15 * 60 * 1000)
      expect(presetMap['30m']).toBe(30 * 60 * 1000)
      expect(presetMap['1h']).toBe(60 * 60 * 1000)
      expect(presetMap['6h']).toBe(6 * 60 * 60 * 1000)
      expect(presetMap['24h']).toBe(24 * 60 * 60 * 1000)
      expect(presetMap['7d']).toBe(7 * 24 * 60 * 60 * 1000)
    })
  })

  describe('REFRESH_INTERVALS', () => {
    it('should have all expected intervals', () => {
      expect(REFRESH_INTERVALS).toHaveLength(6)
      expect(REFRESH_INTERVALS.map(r => r.value)).toEqual([
        'off', '5s', '15s', '30s', '1m', '5m'
      ])
    })

    it('should have correct intervals in milliseconds', () => {
      const intervalMap = Object.fromEntries(REFRESH_INTERVALS.map(r => [r.value, r.interval]))
      expect(intervalMap['off']).toBe(0)
      expect(intervalMap['5s']).toBe(5000)
      expect(intervalMap['15s']).toBe(15000)
      expect(intervalMap['30s']).toBe(30000)
      expect(intervalMap['1m']).toBe(60000)
    })
  })

  describe('timeRange', () => {
    it('should default to 1 hour range', () => {
      const { timeRange } = useTimeRange()
      const now = Date.now()
      const oneHour = 60 * 60 * 1000

      expect(timeRange.value.end).toBe(now)
      expect(timeRange.value.start).toBe(now - oneHour)
    })

    it('should calculate correct range for Last 5 minutes preset', () => {
      const { setPreset, timeRange } = useTimeRange()
      setPreset('5m')

      const now = Date.now()
      const fiveMinutes = 5 * 60 * 1000

      expect(timeRange.value.end).toBe(now)
      expect(timeRange.value.start).toBe(now - fiveMinutes)
    })

    it('should calculate correct range for Last 7 days preset', () => {
      const { setPreset, timeRange } = useTimeRange()
      setPreset('7d')

      const now = Date.now()
      const sevenDays = 7 * 24 * 60 * 60 * 1000

      expect(timeRange.value.end).toBe(now)
      expect(timeRange.value.start).toBe(now - sevenDays)
    })
  })

  describe('setPreset', () => {
    it('should update selectedPreset', () => {
      const { setPreset, selectedPreset } = useTimeRange()

      setPreset('15m')
      expect(selectedPreset.value).toBe('15m')

      setPreset('6h')
      expect(selectedPreset.value).toBe('6h')
    })

    it('should clear custom range when setting preset', () => {
      const { setCustomRange, setPreset, isCustomRange } = useTimeRange()

      setCustomRange(Date.now() - 1000, Date.now())
      expect(isCustomRange.value).toBe(true)

      setPreset('1h')
      expect(isCustomRange.value).toBe(false)
    })

    it('should not update for invalid preset', () => {
      const { setPreset, selectedPreset } = useTimeRange()
      const originalValue = selectedPreset.value

      setPreset('invalid')
      expect(selectedPreset.value).toBe(originalValue)
    })
  })

  describe('setCustomRange', () => {
    it('should set custom range', () => {
      const { setCustomRange, customRange, isCustomRange, timeRange } = useTimeRange()
      const start = Date.now() - 2 * 60 * 60 * 1000
      const end = Date.now()

      setCustomRange(start, end)

      expect(isCustomRange.value).toBe(true)
      expect(customRange.value).toEqual({ start, end })
      expect(timeRange.value).toEqual({ start, end })
    })
  })

  describe('displayText', () => {
    it('should show preset label for preset ranges', () => {
      const { setPreset, displayText } = useTimeRange()

      setPreset('5m')
      expect(displayText.value).toBe('Last 5 minutes')

      setPreset('24h')
      expect(displayText.value).toBe('Last 24 hours')
    })

    it('should show date range for custom ranges', () => {
      const { setCustomRange, displayText } = useTimeRange()

      // Set a custom range
      const start = new Date('2026-02-01T10:30:00Z').getTime()
      const end = new Date('2026-02-02T14:45:00Z').getTime()
      setCustomRange(start, end)

      // The display text should contain date-time values
      expect(displayText.value).toContain('2026-02-01')
      expect(displayText.value).toContain('2026-02-02')
    })
  })

  describe('refresh', () => {
    it('should update lastRefreshTime', () => {
      const { refresh, lastRefreshTime } = useTimeRange()
      const initialTime = lastRefreshTime.value

      vi.advanceTimersByTime(1000)
      refresh()

      expect(lastRefreshTime.value).toBeGreaterThan(initialTime)
    })

    it('should call registered callbacks', () => {
      const { refresh, onRefresh } = useTimeRange()
      const callback = vi.fn()

      onRefresh(callback)
      refresh()

      expect(callback).toHaveBeenCalledTimes(1)
    })

    it('should call multiple registered callbacks', () => {
      const { refresh, onRefresh } = useTimeRange()
      const callback1 = vi.fn()
      const callback2 = vi.fn()

      onRefresh(callback1)
      onRefresh(callback2)
      refresh()

      expect(callback1).toHaveBeenCalledTimes(1)
      expect(callback2).toHaveBeenCalledTimes(1)
    })
  })

  describe('onRefresh', () => {
    it('should return unsubscribe function', () => {
      const { refresh, onRefresh } = useTimeRange()
      const callback = vi.fn()

      const unsubscribe = onRefresh(callback)
      refresh()
      expect(callback).toHaveBeenCalledTimes(1)

      unsubscribe()
      refresh()
      expect(callback).toHaveBeenCalledTimes(1) // Should not be called again
    })
  })

  describe('setRefreshInterval', () => {
    it('should update refresh interval', () => {
      const { setRefreshInterval, refreshIntervalValue } = useTimeRange()

      setRefreshInterval('5s')
      expect(refreshIntervalValue.value).toBe('5s')

      setRefreshInterval('1m')
      expect(refreshIntervalValue.value).toBe('1m')
    })

    it('should not update for invalid interval', () => {
      const { setRefreshInterval, refreshIntervalValue } = useTimeRange()
      const originalValue = refreshIntervalValue.value

      setRefreshInterval('invalid')
      expect(refreshIntervalValue.value).toBe(originalValue)
    })
  })

  describe('auto-refresh', () => {
    it('should trigger callbacks at specified interval', () => {
      const { setRefreshInterval, onRefresh } = useTimeRange()
      const callback = vi.fn()

      onRefresh(callback)
      setRefreshInterval('5s')

      // Should not be called immediately
      expect(callback).not.toHaveBeenCalled()

      // Advance time by 5 seconds
      vi.advanceTimersByTime(5000)
      expect(callback).toHaveBeenCalledTimes(1)

      // Advance time by another 5 seconds
      vi.advanceTimersByTime(5000)
      expect(callback).toHaveBeenCalledTimes(2)
    })

    it('should stop refresh when set to off', () => {
      const { setRefreshInterval, onRefresh } = useTimeRange()
      const callback = vi.fn()

      onRefresh(callback)
      setRefreshInterval('5s')

      vi.advanceTimersByTime(5000)
      expect(callback).toHaveBeenCalledTimes(1)

      setRefreshInterval('off')
      vi.advanceTimersByTime(5000)
      expect(callback).toHaveBeenCalledTimes(1) // Should not be called again
    })

    it('should change interval when updated', () => {
      const { setRefreshInterval, onRefresh } = useTimeRange()
      const callback = vi.fn()

      onRefresh(callback)
      setRefreshInterval('5s')

      vi.advanceTimersByTime(5000)
      expect(callback).toHaveBeenCalledTimes(1)

      // Change to 15s interval
      setRefreshInterval('15s')

      vi.advanceTimersByTime(5000)
      expect(callback).toHaveBeenCalledTimes(1) // Not yet (only 5s of 15s)

      vi.advanceTimersByTime(10000)
      expect(callback).toHaveBeenCalledTimes(2) // Now 15s has passed
    })
  })

  describe('cleanup', () => {
    it('should clear callbacks and stop auto-refresh', () => {
      const { setRefreshInterval, onRefresh, cleanup, refresh } = useTimeRange()
      const callback = vi.fn()

      onRefresh(callback)
      setRefreshInterval('5s')

      cleanup()

      // Manual refresh should not call callback
      refresh()
      expect(callback).not.toHaveBeenCalled()

      // Auto-refresh should not trigger
      vi.advanceTimersByTime(5000)
      expect(callback).not.toHaveBeenCalled()
    })
  })

  describe('pauseAutoRefresh', () => {
    it('should stop auto-refresh when paused', () => {
      const { setRefreshInterval, onRefresh, pauseAutoRefresh, isPaused } = useTimeRange()
      const callback = vi.fn()

      onRefresh(callback)
      setRefreshInterval('5s')

      // First tick should work
      vi.advanceTimersByTime(5000)
      expect(callback).toHaveBeenCalledTimes(1)

      // Pause and verify isPaused is true
      pauseAutoRefresh()
      expect(isPaused.value).toBe(true)

      // Next tick should not trigger callback
      vi.advanceTimersByTime(5000)
      expect(callback).toHaveBeenCalledTimes(1) // Still 1
    })
  })

  describe('resumeAutoRefresh', () => {
    it('should resume auto-refresh after pause', () => {
      const { setRefreshInterval, onRefresh, pauseAutoRefresh, resumeAutoRefresh, isPaused } = useTimeRange()
      const callback = vi.fn()

      onRefresh(callback)
      setRefreshInterval('5s')

      // First tick
      vi.advanceTimersByTime(5000)
      expect(callback).toHaveBeenCalledTimes(1)

      // Pause
      pauseAutoRefresh()
      expect(isPaused.value).toBe(true)
      vi.advanceTimersByTime(5000)
      expect(callback).toHaveBeenCalledTimes(1) // Still 1

      // Resume
      resumeAutoRefresh()
      expect(isPaused.value).toBe(false)

      // Next tick should work
      vi.advanceTimersByTime(5000)
      expect(callback).toHaveBeenCalledTimes(2)
    })

    it('should not start auto-refresh if interval is off', () => {
      const { setRefreshInterval, onRefresh, pauseAutoRefresh, resumeAutoRefresh } = useTimeRange()
      const callback = vi.fn()

      onRefresh(callback)
      setRefreshInterval('off')

      pauseAutoRefresh()
      resumeAutoRefresh()

      vi.advanceTimersByTime(5000)
      expect(callback).not.toHaveBeenCalled()
    })
  })
})
