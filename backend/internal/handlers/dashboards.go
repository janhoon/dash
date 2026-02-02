package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/janhoon/dash/backend/internal/auth"
	"github.com/janhoon/dash/backend/internal/models"
)

type DashboardHandler struct {
	pool *pgxpool.Pool
}

func NewDashboardHandler(pool *pgxpool.Pool) *DashboardHandler {
	return &DashboardHandler{pool: pool}
}

// checkOrgMembership verifies the user is a member of the organization
func (h *DashboardHandler) checkOrgMembership(ctx context.Context, userID, orgID uuid.UUID) (string, error) {
	var role string
	err := h.pool.QueryRow(ctx,
		`SELECT role FROM organization_memberships WHERE user_id = $1 AND organization_id = $2`,
		userID, orgID,
	).Scan(&role)
	return role, err
}

func (h *DashboardHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Get user from auth context
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get organization ID from URL path
	orgIDStr := r.PathValue("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid organization id"}`, http.StatusBadRequest)
		return
	}

	var req models.CreateDashboardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Verify user is member of org
	role, err := h.checkOrgMembership(ctx, userID, orgID)
	if err != nil {
		http.Error(w, `{"error":"not a member of this organization"}`, http.StatusForbidden)
		return
	}

	// Only admin and editor can create dashboards
	if role == "viewer" {
		http.Error(w, `{"error":"viewers cannot create dashboards"}`, http.StatusForbidden)
		return
	}

	var dashboard models.Dashboard
	err = h.pool.QueryRow(ctx,
		`INSERT INTO dashboards (title, description, organization_id, created_by)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, title, description, created_at, updated_at, organization_id, created_by`,
		req.Title, req.Description, orgID, userID,
	).Scan(&dashboard.ID, &dashboard.Title, &dashboard.Description,
		&dashboard.CreatedAt, &dashboard.UpdatedAt, &dashboard.OrganizationID, &dashboard.CreatedBy)

	if err != nil {
		http.Error(w, `{"error":"failed to create dashboard"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dashboard)
}

func (h *DashboardHandler) List(w http.ResponseWriter, r *http.Request) {
	// Get user from auth context
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get organization ID from URL path
	orgIDStr := r.PathValue("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid organization id"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Verify user is member of org
	_, err = h.checkOrgMembership(ctx, userID, orgID)
	if err != nil {
		http.Error(w, `{"error":"not a member of this organization"}`, http.StatusForbidden)
		return
	}

	rows, err := h.pool.Query(ctx,
		`SELECT id, title, description, created_at, updated_at, organization_id, created_by
		 FROM dashboards
		 WHERE organization_id = $1
		 ORDER BY created_at DESC`, orgID)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch dashboards"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	dashboards := []models.Dashboard{}
	for rows.Next() {
		var d models.Dashboard
		if err := rows.Scan(&d.ID, &d.Title, &d.Description, &d.CreatedAt, &d.UpdatedAt, &d.OrganizationID, &d.CreatedBy); err != nil {
			http.Error(w, `{"error":"failed to scan dashboard"}`, http.StatusInternalServerError)
			return
		}
		dashboards = append(dashboards, d)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, `{"error":"failed to iterate dashboards"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboards)
}

func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Get user from auth context
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid dashboard id"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var dashboard models.Dashboard
	err = h.pool.QueryRow(ctx,
		`SELECT id, title, description, created_at, updated_at, organization_id, created_by
		 FROM dashboards
		 WHERE id = $1`, id,
	).Scan(&dashboard.ID, &dashboard.Title, &dashboard.Description,
		&dashboard.CreatedAt, &dashboard.UpdatedAt, &dashboard.OrganizationID, &dashboard.CreatedBy)

	if err != nil {
		http.Error(w, `{"error":"dashboard not found"}`, http.StatusNotFound)
		return
	}

	// Verify user is member of the dashboard's org
	if dashboard.OrganizationID != nil {
		_, err = h.checkOrgMembership(ctx, userID, *dashboard.OrganizationID)
		if err != nil {
			http.Error(w, `{"error":"not a member of this organization"}`, http.StatusForbidden)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

func (h *DashboardHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Get user from auth context
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid dashboard id"}`, http.StatusBadRequest)
		return
	}

	var req models.UpdateDashboardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// First get the dashboard to check org membership
	var orgID *uuid.UUID
	err = h.pool.QueryRow(ctx, `SELECT organization_id FROM dashboards WHERE id = $1`, id).Scan(&orgID)
	if err != nil {
		http.Error(w, `{"error":"dashboard not found"}`, http.StatusNotFound)
		return
	}

	// Verify user is member of the dashboard's org
	if orgID != nil {
		role, err := h.checkOrgMembership(ctx, userID, *orgID)
		if err != nil {
			http.Error(w, `{"error":"not a member of this organization"}`, http.StatusForbidden)
			return
		}
		// Only admin and editor can update
		if role == "viewer" {
			http.Error(w, `{"error":"viewers cannot update dashboards"}`, http.StatusForbidden)
			return
		}
	}

	var dashboard models.Dashboard
	err = h.pool.QueryRow(ctx,
		`UPDATE dashboards
		 SET title = COALESCE($1, title),
		     description = COALESCE($2, description),
		     updated_at = NOW()
		 WHERE id = $3
		 RETURNING id, title, description, created_at, updated_at, organization_id, created_by`,
		req.Title, req.Description, id,
	).Scan(&dashboard.ID, &dashboard.Title, &dashboard.Description,
		&dashboard.CreatedAt, &dashboard.UpdatedAt, &dashboard.OrganizationID, &dashboard.CreatedBy)

	if err != nil {
		http.Error(w, `{"error":"dashboard not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

func (h *DashboardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Get user from auth context
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid dashboard id"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// First get the dashboard to check org membership
	var orgID *uuid.UUID
	err = h.pool.QueryRow(ctx, `SELECT organization_id FROM dashboards WHERE id = $1`, id).Scan(&orgID)
	if err != nil {
		http.Error(w, `{"error":"dashboard not found"}`, http.StatusNotFound)
		return
	}

	// Verify user is member of the dashboard's org
	if orgID != nil {
		role, err := h.checkOrgMembership(ctx, userID, *orgID)
		if err != nil {
			http.Error(w, `{"error":"not a member of this organization"}`, http.StatusForbidden)
			return
		}
		// Only admin can delete
		if role != "admin" {
			http.Error(w, `{"error":"only admins can delete dashboards"}`, http.StatusForbidden)
			return
		}
	}

	result, err := h.pool.Exec(ctx, `DELETE FROM dashboards WHERE id = $1`, id)
	if err != nil {
		http.Error(w, `{"error":"failed to delete dashboard"}`, http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, `{"error":"dashboard not found"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
