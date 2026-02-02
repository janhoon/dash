import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { listDashboards, getDashboard, createDashboard, updateDashboard, deleteDashboard } from './dashboards'

describe('Dashboard API', () => {
  const mockFetch = vi.fn()
  const originalFetch = global.fetch

  beforeEach(() => {
    global.fetch = mockFetch
    mockFetch.mockClear()
  })

  afterEach(() => {
    global.fetch = originalFetch
  })

  describe('listDashboards', () => {
    it('fetches dashboards from API', async () => {
      const mockData = [{ id: '1', title: 'Test' }]
      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockData)
      })

      const result = await listDashboards()
      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/api/dashboards')
      expect(result).toEqual(mockData)
    })

    it('throws error on failure', async () => {
      mockFetch.mockResolvedValue({ ok: false })
      await expect(listDashboards()).rejects.toThrow('Failed to fetch dashboards')
    })
  })

  describe('getDashboard', () => {
    it('fetches single dashboard from API', async () => {
      const mockData = { id: '1', title: 'Test' }
      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockData)
      })

      const result = await getDashboard('1')
      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/api/dashboards/1')
      expect(result).toEqual(mockData)
    })

    it('throws error when not found', async () => {
      mockFetch.mockResolvedValue({ ok: false })
      await expect(getDashboard('1')).rejects.toThrow('Dashboard not found')
    })
  })

  describe('createDashboard', () => {
    it('creates dashboard via API', async () => {
      const mockData = { id: '1', title: 'New Dashboard' }
      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockData)
      })

      const result = await createDashboard({ title: 'New Dashboard' })
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/dashboards',
        expect.objectContaining({
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ title: 'New Dashboard' })
        })
      )
      expect(result).toEqual(mockData)
    })

    it('throws error on failure', async () => {
      mockFetch.mockResolvedValue({ ok: false })
      await expect(createDashboard({ title: 'Test' })).rejects.toThrow('Failed to create dashboard')
    })
  })

  describe('updateDashboard', () => {
    it('updates dashboard via API', async () => {
      const mockData = { id: '1', title: 'Updated' }
      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockData)
      })

      const result = await updateDashboard('1', { title: 'Updated' })
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/dashboards/1',
        expect.objectContaining({
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ title: 'Updated' })
        })
      )
      expect(result).toEqual(mockData)
    })

    it('throws error on failure', async () => {
      mockFetch.mockResolvedValue({ ok: false })
      await expect(updateDashboard('1', { title: 'Test' })).rejects.toThrow('Failed to update dashboard')
    })
  })

  describe('deleteDashboard', () => {
    it('deletes dashboard via API', async () => {
      mockFetch.mockResolvedValue({ ok: true })

      await deleteDashboard('1')
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/dashboards/1',
        expect.objectContaining({ method: 'DELETE' })
      )
    })

    it('throws error on failure', async () => {
      mockFetch.mockResolvedValue({ ok: false })
      await expect(deleteDashboard('1')).rejects.toThrow('Failed to delete dashboard')
    })
  })
})
