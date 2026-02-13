package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/pkg/apperror"
)

const entityColumns = `id, name, type, COALESCE(fields, '{}'::jsonb), owner_id, contact_user_id, created_at, updated_at`

// EntityRepository defines operations for managing entities.
type EntityRepository interface {
	Create(ctx context.Context, input model.CreateEntityInput) (*model.Entity, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Entity, error)
	Update(ctx context.Context, id uuid.UUID, input model.UpdateEntityInput) (*model.Entity, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*model.Entity, error)
	ListByOwnerWithFilters(ctx context.Context, ownerID uuid.UUID, entityType string, limit, offset int) ([]*model.EntityListItem, int, error)
	Search(ctx context.Context, ownerID uuid.UUID, query string) ([]*model.Entity, error)
	ListTypes(ctx context.Context, ownerID uuid.UUID) ([]string, error)
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
	fieldsJSON, err := json.Marshal(input.Fields)
	if err != nil {
		fieldsJSON = []byte("{}")
	}

	var entity model.Entity
	var fieldsRaw []byte
	err = r.db.QueryRow(ctx,
		`INSERT INTO entities (name, type, fields, owner_id, contact_user_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING `+entityColumns,
		input.Name, input.Type, fieldsJSON, input.OwnerID, input.ContactUserID,
	).Scan(
		&entity.ID, &entity.Name, &entity.Type, &fieldsRaw, &entity.OwnerID,
		&entity.ContactUserID, &entity.CreatedAt, &entity.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create entity: %w", err)
	}

	_ = json.Unmarshal(fieldsRaw, &entity.Fields)
	if entity.Fields == nil {
		entity.Fields = make(map[string]string)
	}

	return &entity, nil
}

func (r *pgEntityRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Entity, error) {
	var entity model.Entity
	var fieldsRaw []byte
	err := r.db.QueryRow(ctx,
		`SELECT `+entityColumns+` FROM entities WHERE id = $1`, id,
	).Scan(
		&entity.ID, &entity.Name, &entity.Type, &fieldsRaw, &entity.OwnerID,
		&entity.ContactUserID, &entity.CreatedAt, &entity.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("entity", id.String())
		}
		return nil, fmt.Errorf("find entity by id: %w", err)
	}

	_ = json.Unmarshal(fieldsRaw, &entity.Fields)
	if entity.Fields == nil {
		entity.Fields = make(map[string]string)
	}

	return &entity, nil
}

func (r *pgEntityRepository) Update(ctx context.Context, id uuid.UUID, input model.UpdateEntityInput) (*model.Entity, error) {
	// Build dynamic update
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{id}
	argIdx := 2

	if input.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *input.Name)
		argIdx++
	}
	if input.Type != nil {
		setClauses = append(setClauses, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, *input.Type)
		argIdx++
	}
	if input.Fields != nil {
		fieldsJSON, err := json.Marshal(*input.Fields)
		if err != nil {
			return nil, fmt.Errorf("marshal fields: %w", err)
		}
		setClauses = append(setClauses, fmt.Sprintf("fields = $%d", argIdx))
		args = append(args, fieldsJSON)
		argIdx++
	}

	query := fmt.Sprintf("UPDATE entities SET %s WHERE id = $1 RETURNING %s",
		joinStrings(setClauses, ", "), entityColumns)

	var entity model.Entity
	var fieldsRaw []byte
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&entity.ID, &entity.Name, &entity.Type, &fieldsRaw, &entity.OwnerID,
		&entity.ContactUserID, &entity.CreatedAt, &entity.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("entity", id.String())
		}
		return nil, fmt.Errorf("update entity: %w", err)
	}

	_ = json.Unmarshal(fieldsRaw, &entity.Fields)
	if entity.Fields == nil {
		entity.Fields = make(map[string]string)
	}

	return &entity, nil
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func (r *pgEntityRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*model.Entity, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+entityColumns+` FROM entities WHERE owner_id = $1 ORDER BY name`, ownerID,
	)
	if err != nil {
		return nil, fmt.Errorf("list entities by owner: %w", err)
	}
	defer rows.Close()

	var entities []*model.Entity
	for rows.Next() {
		e, err := scanEntity(rows)
		if err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}
	return entities, rows.Err()
}

func (r *pgEntityRepository) ListByOwnerWithFilters(ctx context.Context, ownerID uuid.UUID, entityType string, limit, offset int) ([]*model.EntityListItem, int, error) {
	// Count query
	countQuery := `SELECT COUNT(*) FROM entities WHERE owner_id = $1`
	countArgs := []interface{}{ownerID}

	listQuery := `SELECT e.` + entityColumns + `, COUNT(de.document_id) as doc_count
		FROM entities e
		LEFT JOIN document_entities de ON e.id = de.entity_id
		WHERE e.owner_id = $1`
	listArgs := []interface{}{ownerID}

	argIdx := 2
	if entityType != "" {
		filter := fmt.Sprintf(" AND type = $%d", argIdx)
		countQuery += filter
		countArgs = append(countArgs, entityType)
		listQuery += fmt.Sprintf(" AND e.type = $%d", argIdx)
		listArgs = append(listArgs, entityType)
		argIdx++
	}

	listQuery += fmt.Sprintf(` GROUP BY e.id ORDER BY e.name LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	listArgs = append(listArgs, limit, offset)

	var total int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count entities: %w", err)
	}

	rows, err := r.db.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list entities with filters: %w", err)
	}
	defer rows.Close()

	var items []*model.EntityListItem
	for rows.Next() {
		var item model.EntityListItem
		var fieldsRaw []byte
		if err := rows.Scan(
			&item.ID, &item.Name, &item.Type, &fieldsRaw, &item.OwnerID,
			&item.ContactUserID, &item.CreatedAt, &item.UpdatedAt,
			&item.DocumentCount,
		); err != nil {
			return nil, 0, fmt.Errorf("scan entity list item: %w", err)
		}
		_ = json.Unmarshal(fieldsRaw, &item.Fields)
		if item.Fields == nil {
			item.Fields = make(map[string]string)
		}
		items = append(items, &item)
	}

	return items, total, rows.Err()
}

func (r *pgEntityRepository) ListTypes(ctx context.Context, ownerID uuid.UUID) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT DISTINCT type FROM entities WHERE owner_id = $1 AND type IS NOT NULL AND type != '' ORDER BY type`,
		ownerID,
	)
	if err != nil {
		return nil, fmt.Errorf("list entity types: %w", err)
	}
	defer rows.Close()

	var types []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("scan entity type: %w", err)
		}
		types = append(types, t)
	}
	return types, rows.Err()
}

func (r *pgEntityRepository) Search(ctx context.Context, ownerID uuid.UUID, query string) ([]*model.Entity, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+entityColumns+`
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
		e, err := scanEntity(rows)
		if err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}
	return entities, rows.Err()
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
		`SELECT e.`+entityColumns+`
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
		e, err := scanEntity(rows)
		if err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}
	return entities, rows.Err()
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

// scanEntity scans a single entity row (from pgx.Rows).
func scanEntity(rows pgx.Rows) (*model.Entity, error) {
	var e model.Entity
	var fieldsRaw []byte
	if err := rows.Scan(
		&e.ID, &e.Name, &e.Type, &fieldsRaw, &e.OwnerID,
		&e.ContactUserID, &e.CreatedAt, &e.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan entity: %w", err)
	}
	_ = json.Unmarshal(fieldsRaw, &e.Fields)
	if e.Fields == nil {
		e.Fields = make(map[string]string)
	}
	return &e, nil
}
