package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/janhoon/dash/backend/internal/models"
)

type DashboardHandler struct {
	pool *pgxpool.Pool
}

func NewDashboardHandler(pool *pgxpool.Pool) *DashboardHandler {
	return &DashboardHandler{pool: pool}
}

func (h *DashboardHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var dashboard models.Dashboard
	err := h.pool.QueryRow(ctx,
		`INSERT INTO dashboards (title, description, user_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, title, description, created_at, updated_at, user_id`,
		req.Title, req.Description, req.UserID,
	).Scan(&dashboard.ID, &dashboard.Title, &dashboard.Description,
		&dashboard.CreatedAt, &dashboard.UpdatedAt, &dashboard.UserID)

	if err != nil {
		http.Error(w, `{"error":"failed to create dashboard"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dashboard)
}

func (h *DashboardHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := h.pool.Query(ctx,
		`SELECT id, title, description, created_at, updated_at, user_id
		 FROM dashboards
		 ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch dashboards"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	dashboards := []models.Dashboard{}
	for rows.Next() {
		var d models.Dashboard
		if err := rows.Scan(&d.ID, &d.Title, &d.Description, &d.CreatedAt, &d.UpdatedAt, &d.UserID); err != nil {
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
		`SELECT id, title, description, created_at, updated_at, user_id
		 FROM dashboards
		 WHERE id = $1`, id,
	).Scan(&dashboard.ID, &dashboard.Title, &dashboard.Description,
		&dashboard.CreatedAt, &dashboard.UpdatedAt, &dashboard.UserID)

	if err != nil {
		http.Error(w, `{"error":"dashboard not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

func (h *DashboardHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var dashboard models.Dashboard
	err = h.pool.QueryRow(ctx,
		`UPDATE dashboards
		 SET title = COALESCE($1, title),
		     description = COALESCE($2, description),
		     updated_at = NOW()
		 WHERE id = $3
		 RETURNING id, title, description, created_at, updated_at, user_id`,
		req.Title, req.Description, id,
	).Scan(&dashboard.ID, &dashboard.Title, &dashboard.Description,
		&dashboard.CreatedAt, &dashboard.UpdatedAt, &dashboard.UserID)

	if err != nil {
		http.Error(w, `{"error":"dashboard not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

func (h *DashboardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid dashboard id"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

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
