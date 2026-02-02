import type {
  Organization,
  CreateOrganizationRequest,
  UpdateOrganizationRequest,
  Member,
  Invitation,
  CreateInvitationRequest,
  UpdateMemberRoleRequest,
} from '../types/organization'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

function getAuthHeaders(): HeadersInit {
  const token = localStorage.getItem('access_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

export async function listOrganizations(): Promise<Organization[]> {
  const response = await fetch(`${API_BASE}/api/orgs`, {
    headers: getAuthHeaders(),
  })
  if (!response.ok) {
    throw new Error('Failed to fetch organizations')
  }
  return response.json()
}

export async function getOrganization(id: string): Promise<Organization> {
  const response = await fetch(`${API_BASE}/api/orgs/${id}`, {
    headers: getAuthHeaders(),
  })
  if (!response.ok) {
    if (response.status === 403) {
      throw new Error('Not a member of this organization')
    }
    throw new Error('Organization not found')
  }
  return response.json()
}

export async function createOrganization(data: CreateOrganizationRequest): Promise<Organization> {
  const response = await fetch(`${API_BASE}/api/orgs`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!response.ok) {
    if (response.status === 409) {
      throw new Error('Organization slug already exists')
    }
    const error = await response.json().catch(() => ({}))
    throw new Error(error.error || 'Failed to create organization')
  }
  return response.json()
}

export async function updateOrganization(
  id: string,
  data: UpdateOrganizationRequest
): Promise<Organization> {
  const response = await fetch(`${API_BASE}/api/orgs/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!response.ok) {
    if (response.status === 403) {
      throw new Error('Admin access required')
    }
    if (response.status === 409) {
      throw new Error('Organization slug already exists')
    }
    throw new Error('Failed to update organization')
  }
  return response.json()
}

export async function deleteOrganization(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/api/orgs/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  })
  if (!response.ok) {
    if (response.status === 403) {
      throw new Error('Admin access required')
    }
    throw new Error('Failed to delete organization')
  }
}

export async function createInvitation(
  orgId: string,
  data: CreateInvitationRequest
): Promise<Invitation> {
  const response = await fetch(`${API_BASE}/api/orgs/${orgId}/invitations`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!response.ok) {
    if (response.status === 403) {
      throw new Error('Admin access required')
    }
    if (response.status === 409) {
      throw new Error('User is already a member')
    }
    const error = await response.json().catch(() => ({}))
    throw new Error(error.error || 'Failed to create invitation')
  }
  return response.json()
}

export async function acceptInvitation(token: string): Promise<Member> {
  const response = await fetch(`${API_BASE}/api/invitations/${token}/accept`, {
    method: 'POST',
    headers: getAuthHeaders(),
  })
  if (!response.ok) {
    if (response.status === 404) {
      throw new Error('Invitation not found or expired')
    }
    if (response.status === 403) {
      throw new Error('Invitation is for a different email address')
    }
    throw new Error('Failed to accept invitation')
  }
  return response.json()
}

export async function listMembers(orgId: string): Promise<Member[]> {
  const response = await fetch(`${API_BASE}/api/orgs/${orgId}/members`, {
    headers: getAuthHeaders(),
  })
  if (!response.ok) {
    if (response.status === 403) {
      throw new Error('Not a member of this organization')
    }
    throw new Error('Failed to fetch members')
  }
  return response.json()
}

export async function updateMemberRole(
  orgId: string,
  userId: string,
  data: UpdateMemberRoleRequest
): Promise<void> {
  const response = await fetch(`${API_BASE}/api/orgs/${orgId}/members/${userId}/role`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  })
  if (!response.ok) {
    if (response.status === 403) {
      throw new Error('Admin access required')
    }
    if (response.status === 400) {
      const error = await response.json().catch(() => ({}))
      throw new Error(error.error || 'Cannot update role')
    }
    throw new Error('Failed to update member role')
  }
}

export async function removeMember(orgId: string, userId: string): Promise<void> {
  const response = await fetch(`${API_BASE}/api/orgs/${orgId}/members/${userId}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  })
  if (!response.ok) {
    if (response.status === 403) {
      throw new Error('Admin access required')
    }
    if (response.status === 400) {
      const error = await response.json().catch(() => ({}))
      throw new Error(error.error || 'Cannot remove member')
    }
    throw new Error('Failed to remove member')
  }
}
