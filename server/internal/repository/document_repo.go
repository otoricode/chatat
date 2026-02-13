package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/pkg/apperror"
)

// DocumentRepository defines operations for managing documents.
type DocumentRepository interface {
	Create(ctx context.Context, input model.CreateDocumentInput) (*model.Document, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Document, error)
	ListByChat(ctx context.Context, chatID uuid.UUID) ([]*model.Document, error)
	ListByTopic(ctx context.Context, topicID uuid.UUID) ([]*model.Document, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*model.Document, error)
	ListByTag(ctx context.Context, tag string) ([]*model.Document, error)
	ListAccessible(ctx context.Context, userID uuid.UUID) ([]*model.Document, error)
	ListCollaborators(ctx context.Context, docID uuid.UUID) ([]*model.DocumentCollaborator, error)
	ListSigners(ctx context.Context, docID uuid.UUID) ([]*model.DocumentSigner, error)
	ListTags(ctx context.Context, docID uuid.UUID) ([]string, error)
	AddCollaborator(ctx context.Context, docID, userID uuid.UUID, role model.CollaboratorRole) error
	RemoveCollaborator(ctx context.Context, docID, userID uuid.UUID) error
	UpdateCollaboratorRole(ctx context.Context, docID, userID uuid.UUID, role model.CollaboratorRole) error
	AddSigner(ctx context.Context, docID, userID uuid.UUID) error
	RemoveSigner(ctx context.Context, docID, userID uuid.UUID) error
	RecordSignature(ctx context.Context, docID, userID uuid.UUID, name string) error
	Lock(ctx context.Context, docID uuid.UUID, lockedBy model.LockedByType) error
	Unlock(ctx context.Context, docID uuid.UUID) error
	AddTag(ctx context.Context, docID uuid.UUID, tag string) error
	RemoveTag(ctx context.Context, docID uuid.UUID, tag string) error
	Update(ctx context.Context, id uuid.UUID, input model.UpdateDocumentInput) (*model.Document, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type pgDocumentRepository struct {
	db *pgxpool.Pool
}

// NewDocumentRepository creates a new PostgreSQL-backed DocumentRepository.
func NewDocumentRepository(db *pgxpool.Pool) DocumentRepository {
	return &pgDocumentRepository{db: db}
}

func (r *pgDocumentRepository) Create(ctx context.Context, input model.CreateDocumentInput) (*model.Document, error) {
	icon := input.Icon
	if icon == "" {
		icon = "📄"
	}

	var doc model.Document
	err := r.db.QueryRow(ctx,
		`INSERT INTO documents (title, icon, owner_id, chat_id, topic_id, is_standalone)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, title, icon, cover, owner_id, chat_id, topic_id, is_standalone,
		           require_sigs, locked, locked_at, locked_by, created_at, updated_at`,
		input.Title, icon, input.OwnerID, input.ChatID, input.TopicID, input.IsStandalone,
	).Scan(
		&doc.ID, &doc.Title, &doc.Icon, &doc.Cover, &doc.OwnerID,
		&doc.ChatID, &doc.TopicID, &doc.IsStandalone, &doc.RequireSigs,
		&doc.Locked, &doc.LockedAt, &doc.LockedBy, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}

	return &doc, nil
}

func (r *pgDocumentRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Document, error) {
	var doc model.Document
	err := r.db.QueryRow(ctx,
		`SELECT id, title, icon, cover, owner_id, chat_id, topic_id, is_standalone,
		        require_sigs, locked, locked_at, locked_by, created_at, updated_at
		 FROM documents WHERE id = $1`, id,
	).Scan(
		&doc.ID, &doc.Title, &doc.Icon, &doc.Cover, &doc.OwnerID,
		&doc.ChatID, &doc.TopicID, &doc.IsStandalone, &doc.RequireSigs,
		&doc.Locked, &doc.LockedAt, &doc.LockedBy, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("document", id.String())
		}
		return nil, fmt.Errorf("find document by id: %w", err)
	}

	return &doc, nil
}

func (r *pgDocumentRepository) listDocuments(ctx context.Context, query string, arg any) ([]*model.Document, error) {
	rows, err := r.db.Query(ctx, query, arg)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		docs = append(docs, &doc)
	}

	return docs, rows.Err()
}

const documentColumns = `id, title, icon, cover, owner_id, chat_id, topic_id, is_standalone,
		        require_sigs, locked, locked_at, locked_by, created_at, updated_at`

func (r *pgDocumentRepository) ListByChat(ctx context.Context, chatID uuid.UUID) ([]*model.Document, error) {
	docs, err := r.listDocuments(ctx,
		`SELECT `+documentColumns+` FROM documents WHERE chat_id = $1 ORDER BY updated_at DESC`, chatID,
	)
	if err != nil {
		return nil, fmt.Errorf("list documents by chat: %w", err)
	}
	return docs, nil
}

func (r *pgDocumentRepository) ListByTopic(ctx context.Context, topicID uuid.UUID) ([]*model.Document, error) {
	docs, err := r.listDocuments(ctx,
		`SELECT `+documentColumns+` FROM documents WHERE topic_id = $1 ORDER BY updated_at DESC`, topicID,
	)
	if err != nil {
		return nil, fmt.Errorf("list documents by topic: %w", err)
	}
	return docs, nil
}

func (r *pgDocumentRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*model.Document, error) {
	docs, err := r.listDocuments(ctx,
		`SELECT `+documentColumns+` FROM documents WHERE owner_id = $1 ORDER BY updated_at DESC`, ownerID,
	)
	if err != nil {
		return nil, fmt.Errorf("list documents by owner: %w", err)
	}
	return docs, nil
}

func (r *pgDocumentRepository) ListAccessible(ctx context.Context, userID uuid.UUID) ([]*model.Document, error) {
	rows, err := r.db.Query(ctx,
		`SELECT DISTINCT d.`+documentColumns+`
		 FROM documents d
		 LEFT JOIN document_collaborators dc ON d.id = dc.document_id
		 WHERE d.owner_id = $1 OR dc.user_id = $1
		 ORDER BY d.updated_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list accessible documents: %w", err)
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
			return nil, fmt.Errorf("scan accessible document: %w", err)
		}
		docs = append(docs, &doc)
	}
	return docs, rows.Err()
}

func (r *pgDocumentRepository) ListCollaborators(ctx context.Context, docID uuid.UUID) ([]*model.DocumentCollaborator, error) {
	rows, err := r.db.Query(ctx,
		`SELECT document_id, user_id, role, added_at
		 FROM document_collaborators WHERE document_id = $1
		 ORDER BY added_at`, docID,
	)
	if err != nil {
		return nil, fmt.Errorf("list collaborators: %w", err)
	}
	defer rows.Close()

	var collaborators []*model.DocumentCollaborator
	for rows.Next() {
		var c model.DocumentCollaborator
		if err := rows.Scan(&c.DocumentID, &c.UserID, &c.Role, &c.AddedAt); err != nil {
			return nil, fmt.Errorf("scan collaborator: %w", err)
		}
		collaborators = append(collaborators, &c)
	}
	return collaborators, rows.Err()
}

func (r *pgDocumentRepository) ListSigners(ctx context.Context, docID uuid.UUID) ([]*model.DocumentSigner, error) {
	rows, err := r.db.Query(ctx,
		`SELECT document_id, user_id, signed_at, signer_name
		 FROM document_signers WHERE document_id = $1`, docID,
	)
	if err != nil {
		return nil, fmt.Errorf("list signers: %w", err)
	}
	defer rows.Close()

	var signers []*model.DocumentSigner
	for rows.Next() {
		var s model.DocumentSigner
		if err := rows.Scan(&s.DocumentID, &s.UserID, &s.SignedAt, &s.SignerName); err != nil {
			return nil, fmt.Errorf("scan signer: %w", err)
		}
		signers = append(signers, &s)
	}
	return signers, rows.Err()
}

func (r *pgDocumentRepository) ListTags(ctx context.Context, docID uuid.UUID) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT tag FROM document_tags WHERE document_id = $1 ORDER BY tag`, docID,
	)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *pgDocumentRepository) UpdateCollaboratorRole(ctx context.Context, docID, userID uuid.UUID, role model.CollaboratorRole) error {
	result, err := r.db.Exec(ctx,
		`UPDATE document_collaborators SET role = $3 WHERE document_id = $1 AND user_id = $2`,
		docID, userID, role,
	)
	if err != nil {
		return fmt.Errorf("update collaborator role: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperror.NotFound("collaborator", userID.String())
	}
	return nil
}

func (r *pgDocumentRepository) ListByTag(ctx context.Context, tag string) ([]*model.Document, error) {
	rows, err := r.db.Query(ctx,
		`SELECT d.`+documentColumns+`
		 FROM documents d
		 JOIN document_tags dt ON d.id = dt.document_id
		 WHERE dt.tag = $1
		 ORDER BY d.updated_at DESC`, tag,
	)
	if err != nil {
		return nil, fmt.Errorf("list documents by tag: %w", err)
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
			return nil, fmt.Errorf("scan document by tag: %w", err)
		}
		docs = append(docs, &doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate document tag rows: %w", err)
	}

	return docs, nil
}

func (r *pgDocumentRepository) AddCollaborator(ctx context.Context, docID, userID uuid.UUID, role model.CollaboratorRole) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO document_collaborators (document_id, user_id, role) VALUES ($1, $2, $3)
		 ON CONFLICT (document_id, user_id) DO UPDATE SET role = $3`,
		docID, userID, role,
	)
	if err != nil {
		return fmt.Errorf("add collaborator: %w", err)
	}

	return nil
}

func (r *pgDocumentRepository) RemoveCollaborator(ctx context.Context, docID, userID uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM document_collaborators WHERE document_id = $1 AND user_id = $2`,
		docID, userID,
	)
	if err != nil {
		return fmt.Errorf("remove collaborator: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("collaborator", userID.String())
	}

	return nil
}

func (r *pgDocumentRepository) AddSigner(ctx context.Context, docID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO document_signers (document_id, user_id) VALUES ($1, $2)
		 ON CONFLICT (document_id, user_id) DO NOTHING`,
		docID, userID,
	)
	if err != nil {
		return fmt.Errorf("add signer: %w", err)
	}

	return nil
}

func (r *pgDocumentRepository) RecordSignature(ctx context.Context, docID, userID uuid.UUID, name string) error {
	result, err := r.db.Exec(ctx,
		`UPDATE document_signers SET signed_at = $3, signer_name = $4
		 WHERE document_id = $1 AND user_id = $2`,
		docID, userID, time.Now(), name,
	)
	if err != nil {
		return fmt.Errorf("record signature: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("signer", userID.String())
	}

	return nil
}

func (r *pgDocumentRepository) Lock(ctx context.Context, docID uuid.UUID, lockedBy model.LockedByType) error {
	result, err := r.db.Exec(ctx,
		`UPDATE documents SET locked = true, locked_at = NOW(), locked_by = $2, updated_at = NOW()
		 WHERE id = $1`,
		docID, lockedBy,
	)
	if err != nil {
		return fmt.Errorf("lock document: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("document", docID.String())
	}

	return nil
}

func (r *pgDocumentRepository) Unlock(ctx context.Context, docID uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`UPDATE documents SET locked = false, locked_at = NULL, locked_by = NULL, updated_at = NOW()
		 WHERE id = $1`,
		docID,
	)
	if err != nil {
		return fmt.Errorf("unlock document: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("document", docID.String())
	}

	return nil
}

func (r *pgDocumentRepository) RemoveSigner(ctx context.Context, docID, userID uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM document_signers WHERE document_id = $1 AND user_id = $2`,
		docID, userID,
	)
	if err != nil {
		return fmt.Errorf("remove signer: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("signer", userID.String())
	}

	return nil
}

func (r *pgDocumentRepository) AddTag(ctx context.Context, docID uuid.UUID, tag string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO document_tags (document_id, tag) VALUES ($1, $2)
		 ON CONFLICT (document_id, tag) DO NOTHING`,
		docID, tag,
	)
	if err != nil {
		return fmt.Errorf("add tag: %w", err)
	}

	return nil
}

func (r *pgDocumentRepository) RemoveTag(ctx context.Context, docID uuid.UUID, tag string) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM document_tags WHERE document_id = $1 AND tag = $2`,
		docID, tag,
	)
	if err != nil {
		return fmt.Errorf("remove tag: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("tag", tag)
	}

	return nil
}

func (r *pgDocumentRepository) Update(ctx context.Context, id uuid.UUID, input model.UpdateDocumentInput) (*model.Document, error) {
	var doc model.Document
	err := r.db.QueryRow(ctx,
		`UPDATE documents SET
		   title = COALESCE($2, title),
		   icon = COALESCE($3, icon),
		   cover = COALESCE($4, cover),
		   require_sigs = COALESCE($5, require_sigs),
		   updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, title, icon, cover, owner_id, chat_id, topic_id, is_standalone,
		           require_sigs, locked, locked_at, locked_by, created_at, updated_at`,
		id, input.Title, input.Icon, input.Cover, input.RequireSigs,
	).Scan(
		&doc.ID, &doc.Title, &doc.Icon, &doc.Cover, &doc.OwnerID,
		&doc.ChatID, &doc.TopicID, &doc.IsStandalone, &doc.RequireSigs,
		&doc.Locked, &doc.LockedAt, &doc.LockedBy, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("document", id.String())
		}
		return nil, fmt.Errorf("update document: %w", err)
	}

	return &doc, nil
}

func (r *pgDocumentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM documents WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("document", id.String())
	}

	return nil
}
