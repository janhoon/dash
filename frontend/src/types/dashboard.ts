export interface Dashboard {
  id: string
  title: string
  description?: string
  created_at: string
  updated_at: string
  user_id?: string
  organization_id?: string
  created_by?: string
}

export interface CreateDashboardRequest {
  title: string
  description?: string
}

export interface UpdateDashboardRequest {
  title?: string
  description?: string
}
