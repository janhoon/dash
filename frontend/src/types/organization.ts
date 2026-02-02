export type MembershipRole = 'admin' | 'editor' | 'viewer'

export interface Organization {
  id: string
  name: string
  slug: string
  created_at: string
  updated_at: string
  role?: MembershipRole
}

export interface CreateOrganizationRequest {
  name: string
  slug: string
}

export interface UpdateOrganizationRequest {
  name?: string
  slug?: string
}

export interface Member {
  id: string
  user_id: string
  email: string
  name?: string
  role: MembershipRole
  created_at: string
}

export interface Invitation {
  token: string
  email: string
  role: MembershipRole
  expires_at: string
}

export interface CreateInvitationRequest {
  email: string
  role: MembershipRole
}

export interface UpdateMemberRoleRequest {
  role: MembershipRole
}
