package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/otoritech/chatat/internal/model"
)

// BackupRepository defines operations for managing backup records.
type BackupRepository interface {
	Create(ctx context.Context, userID uuid.UUID, input model.LogBackupInput) (*model.BackupRecord, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.BackupStatus, sizeBytes int64) error
	FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]model.BackupRecord, error)
	FindLatestByUser(ctx context.Context, userID uuid.UUID) (*model.BackupRecord, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type pgBackupRepository struct {
	db *pgxpool.Pool
}

// NewBackupRepository creates a new PostgreSQL-backed BackupRepository.
func NewBackupRepository(db *pgxpool.Pool) BackupRepository {
	return &pgBackupRepository{db: db}
}

func (r *pgBackupRepository) Create(ctx context.Context, userID uuid.UUID, input model.LogBackupInput) (*model.BackupRecord, error) {
	status := input.Status
	if status == "" {
		status = model.BackupStatusCompleted
	}

	var rec model.BackupRecord
	err := r.db.QueryRow(ctx,
		`INSERT INTO backup_records (user_id, size_bytes, platform, status, metadata)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, size_bytes, platform, status, metadata, created_at`,
		userID, input.SizeBytes, input.Platform, status, json.RawMessage("{}"),
	).Scan(&rec.ID, &rec.UserID, &rec.SizeBytes, &rec.Platform, &rec.Status, &rec.Metadata, &rec.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create backup record: %w", err)
	}
	return &rec, nil
}

func (r *pgBackupRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.BackupStatus, sizeBytes int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE backup_records SET status = $1, size_bytes = $2 WHERE id = $3`,
		status, sizeBytes, id,
	)
	if err != nil {
		return fmt.Errorf("update backup status: %w", err)
	}
	return nil
}

func (r *pgBackupRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]model.BackupRecord, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, size_bytes, platform, status, metadata, created_at
		 FROM backup_records
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("find backup records: %w", err)
	}
	defer rows.Close()

	var records []model.BackupRecord
	for rows.Next() {
		var rec model.BackupRecord
		if err := rows.Scan(&rec.ID, &rec.UserID, &rec.SizeBytes, &rec.Platform, &rec.Status, &rec.Metadata, &rec.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan backup record: %w", err)
		}
		records = append(records, rec)
	}
	return records, nil
}

func (r *pgBackupRepository) FindLatestByUser(ctx context.Context, userID uuid.UUID) (*model.BackupRecord, error) {
	var rec model.BackupRecord
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, size_bytes, platform, status, metadata, created_at
		 FROM backup_records
		 WHERE user_id = $1 AND status = 'completed'
		 ORDER BY created_at DESC
		 LIMIT 1`,
		userID,
	).Scan(&rec.ID, &rec.UserID, &rec.SizeBytes, &rec.Platform, &rec.Status, &rec.Metadata, &rec.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("find latest backup: %w", err)
	}
	return &rec, nil
}

func (r *pgBackupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM backup_records WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete backup record: %w", err)
	}
	return nil
}
