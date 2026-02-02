import type { Dashboard, CreateDashboardRequest, UpdateDashboardRequest } from '../types/dashboard'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

function getAuthHeaders(): HeadersInit {
  const token = localStorage.getItem('access_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

export async function listDashboards(orgId: string): Promise<Dashboard[]> {
  const response = await fetch(`${API_BASE}/api/orgs/${orgId}/dashboards`, {
    headers: getAuthHeaders(),
  })
  if (!response.ok) {
    if (response.status === 403) {
      throw new Error('Not a member of this organization')
    }
    throw new Error('Failed to fetch dashboards')
  }
  return response.json()
}

export async function getDashboard(id: string): Promise<Dashboard> {
  const response = await fetch(`${API_BASE}/api/dashboards/${id}`, {
    headers: getAuthHeaders(),
  })
  if (!response.ok) {
    if (response.status === 403) {
      throw new Error('Not a member of this organization')
    }
    throw new Error('Dashboard not found')
  }
  return response.json()
}

export async function createDashboard(orgId: string, data: CreateDashboardRequest): Promise<Dashboard> {
  const response = await fetch(`${API_BASE}/api/orgs/${orgId}/dashboards`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!response.ok) {
    if (response.status === 403) {
      throw new Error('Not authorized to create dashboards in this organization')
    }
    throw new Error('Failed to create dashboard')
  }
  return response.json()
}

export async function updateDashboard(id: string, data: UpdateDashboardRequest): Promise<Dashboard> {
  const response = await fetch(`${API_BASE}/api/dashboards/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!response.ok) {
    if (response.status === 403) {
      throw new Error('Not authorized to update this dashboard')
    }
    throw new Error('Failed to update dashboard')
  }
  return response.json()
}

export async function deleteDashboard(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/api/dashboards/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  })
  if (!response.ok) {
    if (response.status === 403) {
      throw new Error('Not authorized to delete this dashboard')
    }
    throw new Error('Failed to delete dashboard')
  }
}
