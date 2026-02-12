package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/pkg/apperror"
)

// BlockRepository defines operations for managing document blocks.
type BlockRepository interface {
	Create(ctx context.Context, input model.CreateBlockInput) (*model.Block, error)
	ListByDocument(ctx context.Context, docID uuid.UUID) ([]*model.Block, error)
	Update(ctx context.Context, id uuid.UUID, input model.UpdateBlockInput) (*model.Block, error)
	Reorder(ctx context.Context, docID uuid.UUID, blockIDs []uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type pgBlockRepository struct {
	db *pgxpool.Pool
}

// NewBlockRepository creates a new PostgreSQL-backed BlockRepository.
func NewBlockRepository(db *pgxpool.Pool) BlockRepository {
	return &pgBlockRepository{db: db}
}

func (r *pgBlockRepository) Create(ctx context.Context, input model.CreateBlockInput) (*model.Block, error) {
	var block model.Block
	err := r.db.QueryRow(ctx,
		`INSERT INTO blocks (document_id, type, content, checked, rows, columns, language, emoji, color, sort_order, parent_block_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id, document_id, type, content, checked, rows, columns, language, emoji, color, sort_order, parent_block_id, created_at, updated_at`,
		input.DocumentID, input.Type, input.Content, input.Checked, input.Rows,
		input.Columns, input.Language, input.Emoji, input.Color, input.SortOrder, input.ParentBlockID,
	).Scan(
		&block.ID, &block.DocumentID, &block.Type, &block.Content, &block.Checked,
		&block.Rows, &block.Columns, &block.Language, &block.Emoji, &block.Color,
		&block.SortOrder, &block.ParentBlockID, &block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create block: %w", err)
	}

	return &block, nil
}

func (r *pgBlockRepository) ListByDocument(ctx context.Context, docID uuid.UUID) ([]*model.Block, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, document_id, type, content, checked, rows, columns, language, emoji, color, sort_order, parent_block_id, created_at, updated_at
		 FROM blocks WHERE document_id = $1
		 ORDER BY sort_order`, docID,
	)
	if err != nil {
		return nil, fmt.Errorf("list blocks by document: %w", err)
	}
	defer rows.Close()

	var blocks []*model.Block
	for rows.Next() {
		var b model.Block
		if err := rows.Scan(
			&b.ID, &b.DocumentID, &b.Type, &b.Content, &b.Checked,
			&b.Rows, &b.Columns, &b.Language, &b.Emoji, &b.Color,
			&b.SortOrder, &b.ParentBlockID, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan block row: %w", err)
		}
		blocks = append(blocks, &b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate block rows: %w", err)
	}

	return blocks, nil
}

func (r *pgBlockRepository) Update(ctx context.Context, id uuid.UUID, input model.UpdateBlockInput) (*model.Block, error) {
	var block model.Block
	err := r.db.QueryRow(ctx,
		`UPDATE blocks SET
		   content = COALESCE($2, content),
		   checked = COALESCE($3, checked),
		   rows = COALESCE($4, rows),
		   columns = COALESCE($5, columns),
		   language = COALESCE($6, language),
		   emoji = COALESCE($7, emoji),
		   color = COALESCE($8, color),
		   sort_order = COALESCE($9, sort_order),
		   updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, document_id, type, content, checked, rows, columns, language, emoji, color, sort_order, parent_block_id, created_at, updated_at`,
		id, input.Content, input.Checked, input.Rows, input.Columns,
		input.Language, input.Emoji, input.Color, input.SortOrder,
	).Scan(
		&block.ID, &block.DocumentID, &block.Type, &block.Content, &block.Checked,
		&block.Rows, &block.Columns, &block.Language, &block.Emoji, &block.Color,
		&block.SortOrder, &block.ParentBlockID, &block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("block", id.String())
		}
		return nil, fmt.Errorf("update block: %w", err)
	}

	return &block, nil
}

func (r *pgBlockRepository) Reorder(ctx context.Context, docID uuid.UUID, blockIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin reorder transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for i, blockID := range blockIDs {
		_, err := tx.Exec(ctx,
			`UPDATE blocks SET sort_order = $1, updated_at = NOW()
			 WHERE id = $2 AND document_id = $3`,
			i, blockID, docID,
		)
		if err != nil {
			return fmt.Errorf("reorder block %s: %w", blockID.String(), err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit reorder transaction: %w", err)
	}

	return nil
}

func (r *pgBlockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM blocks WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("delete block: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("block", id.String())
	}

	return nil
}
