package models

import (
	"time"

	"github.com/google/uuid"
)

type DataSource struct {
	ID        uuid.UUID              `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	URL       string                 `json:"url"`
	Config    map[string]interface{} `json:"config,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}
