export interface Dashboard {
  id: string
  title: string
  description?: string
  created_at: string
  updated_at: string
  user_id?: string
}

export interface CreateDashboardRequest {
  title: string
  description?: string
  user_id?: string
}

export interface UpdateDashboardRequest {
  title?: string
  description?: string
}
