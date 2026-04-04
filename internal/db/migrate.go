package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrate создаёт все необходимые таблицы в базе данных
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	slog.Info("Starting database migrations...")

	queries := []struct {
		name  string
		query string
	}{
		{
			name: "create universities table",
			query: `CREATE TABLE IF NOT EXISTS universities (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				code VARCHAR(50) UNIQUE NOT NULL,
				official_name TEXT NOT NULL,
				public_key_pem TEXT NOT NULL,
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW()
			)`,
		},
		{
			name: "create diplomas table",
			query: `CREATE TABLE IF NOT EXISTS diplomas (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				university_id UUID NOT NULL REFERENCES universities(id) ON DELETE CASCADE,
				serial_number VARCHAR(100) NOT NULL,
				encrypted_fio_full TEXT NOT NULL,
				encrypted_passport_sn TEXT,
				specialty VARCHAR(200) NOT NULL,
				issue_year INTEGER NOT NULL CHECK (issue_year >= 1900 AND issue_year <= 2100),
				document_hash VARCHAR(255) NOT NULL,
				digital_signature TEXT NOT NULL,
				status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'revoked', 'expired')),
				revoked_at TIMESTAMPTZ,
				revocation_reason TEXT,
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW(),
				UNIQUE(university_id, serial_number)
			)`,
		},
		{
			name:  "create index diplomas_serial",
			query: `CREATE INDEX IF NOT EXISTS idx_diplomas_serial ON diplomas(serial_number)`,
		},
		{
			name:  "create index diplomas_university",
			query: `CREATE INDEX IF NOT EXISTS idx_diplomas_university ON diplomas(university_id)`,
		},
		{
			name:  "create index diplomas_status",
			query: `CREATE INDEX IF NOT EXISTS idx_diplomas_status ON diplomas(status)`,
		},
		{
			name: "insert test university",
			query: `INSERT INTO universities (id, code, official_name, public_key_pem)
				VALUES ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'TEST01', 'Тестовый ВУЗ', 'stub_public_key')
				ON CONFLICT (code) DO NOTHING`,
		},
	}

	for _, q := range queries {
		if _, err := pool.Exec(ctx, q.query); err != nil {
			return fmt.Errorf("migration %s failed: %w", q.name, err)
		}
		slog.Info("Migration applied", "name", q.name)
	}

	slog.Info("All database migrations applied successfully")
	return nil
}