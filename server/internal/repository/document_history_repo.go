package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/otoritech/chatat/internal/model"
)

// DocumentHistoryRepository defines operations for managing document history.
type DocumentHistoryRepository interface {
	Create(ctx context.Context, docID, userID uuid.UUID, action, details string) error
	ListByDocument(ctx context.Context, docID uuid.UUID) ([]*model.DocumentHistory, error)
}

type pgDocumentHistoryRepository struct {
	db *pgxpool.Pool
}

// NewDocumentHistoryRepository creates a new PostgreSQL-backed DocumentHistoryRepository.
func NewDocumentHistoryRepository(db *pgxpool.Pool) DocumentHistoryRepository {
	return &pgDocumentHistoryRepository{db: db}
}

func (r *pgDocumentHistoryRepository) Create(ctx context.Context, docID, userID uuid.UUID, action, details string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO document_history (document_id, user_id, action, details) VALUES ($1, $2, $3, $4)`,
		docID, userID, action, details,
	)
	if err != nil {
		return fmt.Errorf("create document history: %w", err)
	}

	return nil
}

func (r *pgDocumentHistoryRepository) ListByDocument(ctx context.Context, docID uuid.UUID) ([]*model.DocumentHistory, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, document_id, user_id, action, details, created_at
		 FROM document_history WHERE document_id = $1
		 ORDER BY created_at DESC`, docID,
	)
	if err != nil {
		return nil, fmt.Errorf("list document history: %w", err)
	}
	defer rows.Close()

	var history []*model.DocumentHistory
	for rows.Next() {
		var h model.DocumentHistory
		if err := rows.Scan(&h.ID, &h.DocumentID, &h.UserID, &h.Action, &h.Details, &h.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan document history: %w", err)
		}
		history = append(history, &h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate document history rows: %w", err)
	}

	return history, nil
}
