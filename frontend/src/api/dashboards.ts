import type { Dashboard, CreateDashboardRequest, UpdateDashboardRequest } from '../types/dashboard'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export async function listDashboards(): Promise<Dashboard[]> {
  const response = await fetch(`${API_BASE}/api/dashboards`)
  if (!response.ok) {
    throw new Error('Failed to fetch dashboards')
  }
  return response.json()
}

export async function getDashboard(id: string): Promise<Dashboard> {
  const response = await fetch(`${API_BASE}/api/dashboards/${id}`)
  if (!response.ok) {
    throw new Error('Dashboard not found')
  }
  return response.json()
}

export async function createDashboard(data: CreateDashboardRequest): Promise<Dashboard> {
  const response = await fetch(`${API_BASE}/api/dashboards`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  })
  if (!response.ok) {
    throw new Error('Failed to create dashboard')
  }
  return response.json()
}

export async function updateDashboard(id: string, data: UpdateDashboardRequest): Promise<Dashboard> {
  const response = await fetch(`${API_BASE}/api/dashboards/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  })
  if (!response.ok) {
    throw new Error('Failed to update dashboard')
  }
  return response.json()
}

export async function deleteDashboard(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/api/dashboards/${id}`, {
    method: 'DELETE',
  })
  if (!response.ok) {
    throw new Error('Failed to delete dashboard')
  }
}
