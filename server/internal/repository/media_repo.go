package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/otoritech/chatat/internal/model"
)

// MediaRepository handles media database operations.
type MediaRepository interface {
	Create(ctx context.Context, media *model.Media) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Media, error)
	ListByContext(ctx context.Context, contextType string, contextID uuid.UUID) ([]*model.Media, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type pgMediaRepository struct {
	db *pgxpool.Pool
}

// NewMediaRepository creates a new PostgreSQL-backed media repository.
func NewMediaRepository(db *pgxpool.Pool) MediaRepository {
	return &pgMediaRepository{db: db}
}

func (r *pgMediaRepository) Create(ctx context.Context, media *model.Media) error {
	query := `
		INSERT INTO media (id, uploader_id, type, filename, content_type, size, width, height, storage_key, thumbnail_key, context_type, context_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.Exec(ctx, query,
		media.ID, media.UploaderID, media.Type, media.Filename, media.ContentType,
		media.Size, media.Width, media.Height, media.StorageKey, media.ThumbnailKey,
		media.ContextType, media.ContextID, media.CreatedAt,
	)
	return err
}

func (r *pgMediaRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Media, error) {
	query := `
		SELECT id, uploader_id, type, filename, content_type, size, width, height, storage_key, thumbnail_key, context_type, context_id, created_at
		FROM media WHERE id = $1
	`
	var m model.Media
	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.UploaderID, &m.Type, &m.Filename, &m.ContentType,
		&m.Size, &m.Width, &m.Height, &m.StorageKey, &m.ThumbnailKey,
		&m.ContextType, &m.ContextID, &m.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *pgMediaRepository) ListByContext(ctx context.Context, contextType string, contextID uuid.UUID) ([]*model.Media, error) {
	query := `
		SELECT id, uploader_id, type, filename, content_type, size, width, height, storage_key, thumbnail_key, context_type, context_id, created_at
		FROM media WHERE context_type = $1 AND context_id = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, contextType, contextID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.Media
	for rows.Next() {
		var m model.Media
		if err := rows.Scan(
			&m.ID, &m.UploaderID, &m.Type, &m.Filename, &m.ContentType,
			&m.Size, &m.Width, &m.Height, &m.StorageKey, &m.ThumbnailKey,
			&m.ContextType, &m.ContextID, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, &m)
	}
	return result, rows.Err()
}

func (r *pgMediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM media WHERE id = $1", id)
	return err
}
