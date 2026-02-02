package models

import (
	"time"

	"github.com/google/uuid"
)

type Dashboard struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	UserID      *string    `json:"user_id,omitempty"`
}

type CreateDashboardRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	UserID      *string `json:"user_id,omitempty"`
}

type UpdateDashboardRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}
