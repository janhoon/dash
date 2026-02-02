package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		// Create dashboards table
		`CREATE TABLE IF NOT EXISTS dashboards (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title VARCHAR(255) NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			user_id VARCHAR(255)
		)`,
		// Create panels table
		`CREATE TABLE IF NOT EXISTS panels (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			dashboard_id UUID REFERENCES dashboards(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			type VARCHAR(50) DEFAULT 'line_chart',
			grid_pos JSONB NOT NULL,
			query JSONB,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		// Create data_sources table
		`CREATE TABLE IF NOT EXISTS data_sources (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			url VARCHAR(500) NOT NULL,
			config JSONB,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
	}

	for _, migration := range migrations {
		_, err := pool.Exec(ctx, migration)
		if err != nil {
			return err
		}
	}

	return nil
}
