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

// EntityRepository defines operations for managing entities.
type EntityRepository interface {
	Create(ctx context.Context, input model.CreateEntityInput) (*model.Entity, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Entity, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*model.Entity, error)
	Search(ctx context.Context, ownerID uuid.UUID, query string) ([]*model.Entity, error)
	LinkToDocument(ctx context.Context, docID, entityID uuid.UUID) error
	UnlinkFromDocument(ctx context.Context, docID, entityID uuid.UUID) error
	ListByDocument(ctx context.Context, docID uuid.UUID) ([]*model.Entity, error)
	ListDocumentsByEntity(ctx context.Context, entityID uuid.UUID) ([]*model.Document, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type pgEntityRepository struct {
	db *pgxpool.Pool
}

// NewEntityRepository creates a new PostgreSQL-backed EntityRepository.
func NewEntityRepository(db *pgxpool.Pool) EntityRepository {
	return &pgEntityRepository{db: db}
}

func (r *pgEntityRepository) Create(ctx context.Context, input model.CreateEntityInput) (*model.Entity, error) {
	var entity model.Entity
	err := r.db.QueryRow(ctx,
		`INSERT INTO entities (name, type, owner_id, contact_user_id)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, name, type, owner_id, contact_user_id, created_at`,
		input.Name, input.Type, input.OwnerID, input.ContactUserID,
	).Scan(
		&entity.ID, &entity.Name, &entity.Type, &entity.OwnerID,
		&entity.ContactUserID, &entity.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create entity: %w", err)
	}

	return &entity, nil
}

func (r *pgEntityRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Entity, error) {
	var entity model.Entity
	err := r.db.QueryRow(ctx,
		`SELECT id, name, type, owner_id, contact_user_id, created_at
		 FROM entities WHERE id = $1`, id,
	).Scan(
		&entity.ID, &entity.Name, &entity.Type, &entity.OwnerID,
		&entity.ContactUserID, &entity.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("entity", id.String())
		}
		return nil, fmt.Errorf("find entity by id: %w", err)
	}

	return &entity, nil
}

func (r *pgEntityRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*model.Entity, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, type, owner_id, contact_user_id, created_at
		 FROM entities WHERE owner_id = $1
		 ORDER BY name`, ownerID,
	)
	if err != nil {
		return nil, fmt.Errorf("list entities by owner: %w", err)
	}
	defer rows.Close()

	var entities []*model.Entity
	for rows.Next() {
		var e model.Entity
		if err := rows.Scan(&e.ID, &e.Name, &e.Type, &e.OwnerID, &e.ContactUserID, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan entity row: %w", err)
		}
		entities = append(entities, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity rows: %w", err)
	}

	return entities, nil
}

func (r *pgEntityRepository) Search(ctx context.Context, ownerID uuid.UUID, query string) ([]*model.Entity, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, type, owner_id, contact_user_id, created_at
		 FROM entities
		 WHERE owner_id = $1 AND name ILIKE '%' || $2 || '%'
		 ORDER BY name
		 LIMIT 50`, ownerID, query,
	)
	if err != nil {
		return nil, fmt.Errorf("search entities: %w", err)
	}
	defer rows.Close()

	var entities []*model.Entity
	for rows.Next() {
		var e model.Entity
		if err := rows.Scan(&e.ID, &e.Name, &e.Type, &e.OwnerID, &e.ContactUserID, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan entity search row: %w", err)
		}
		entities = append(entities, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity search rows: %w", err)
	}

	return entities, nil
}

func (r *pgEntityRepository) LinkToDocument(ctx context.Context, docID, entityID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO document_entities (document_id, entity_id) VALUES ($1, $2)
		 ON CONFLICT (document_id, entity_id) DO NOTHING`,
		docID, entityID,
	)
	if err != nil {
		return fmt.Errorf("link entity to document: %w", err)
	}

	return nil
}

func (r *pgEntityRepository) UnlinkFromDocument(ctx context.Context, docID, entityID uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM document_entities WHERE document_id = $1 AND entity_id = $2`,
		docID, entityID,
	)
	if err != nil {
		return fmt.Errorf("unlink entity from document: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("document entity link", entityID.String())
	}

	return nil
}

func (r *pgEntityRepository) ListByDocument(ctx context.Context, docID uuid.UUID) ([]*model.Entity, error) {
	rows, err := r.db.Query(ctx,
		`SELECT e.id, e.name, e.type, e.owner_id, e.contact_user_id, e.created_at
		 FROM entities e
		 JOIN document_entities de ON e.id = de.entity_id
		 WHERE de.document_id = $1
		 ORDER BY e.name`, docID,
	)
	if err != nil {
		return nil, fmt.Errorf("list entities by document: %w", err)
	}
	defer rows.Close()

	var entities []*model.Entity
	for rows.Next() {
		var e model.Entity
		if err := rows.Scan(&e.ID, &e.Name, &e.Type, &e.OwnerID, &e.ContactUserID, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan entity by document: %w", err)
		}
		entities = append(entities, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity document rows: %w", err)
	}

	return entities, nil
}

func (r *pgEntityRepository) ListDocumentsByEntity(ctx context.Context, entityID uuid.UUID) ([]*model.Document, error) {
	rows, err := r.db.Query(ctx,
		`SELECT d.`+documentColumns+`
		 FROM documents d
		 JOIN document_entities de ON d.id = de.document_id
		 WHERE de.entity_id = $1
		 ORDER BY d.updated_at DESC`, entityID,
	)
	if err != nil {
		return nil, fmt.Errorf("list documents by entity: %w", err)
	}
	defer rows.Close()

	var docs []*model.Document
	for rows.Next() {
		var doc model.Document
		if err := rows.Scan(
			&doc.ID, &doc.Title, &doc.Icon, &doc.Cover, &doc.OwnerID,
			&doc.ChatID, &doc.TopicID, &doc.IsStandalone, &doc.RequireSigs,
			&doc.Locked, &doc.LockedAt, &doc.LockedBy, &doc.CreatedAt, &doc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan document by entity: %w", err)
		}
		docs = append(docs, &doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate document entity rows: %w", err)
	}

	return docs, nil
}

func (r *pgEntityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM entities WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("delete entity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("entity", id.String())
	}

	return nil
}
