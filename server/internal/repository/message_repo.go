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

// MessageRepository defines operations for managing chat messages.
type MessageRepository interface {
	Create(ctx context.Context, input model.CreateMessageInput) (*model.Message, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Message, error)
	ListByChat(ctx context.Context, chatID uuid.UUID, cursor *time.Time, limit int) ([]*model.Message, error)
	MarkAsDeleted(ctx context.Context, id uuid.UUID, forAll bool) error
	Search(ctx context.Context, chatID uuid.UUID, query string) ([]*model.Message, error)
}

type pgMessageRepository struct {
	db *pgxpool.Pool
}

// NewMessageRepository creates a new PostgreSQL-backed MessageRepository.
func NewMessageRepository(db *pgxpool.Pool) MessageRepository {
	return &pgMessageRepository{db: db}
}

func (r *pgMessageRepository) Create(ctx context.Context, input model.CreateMessageInput) (*model.Message, error) {
	msgType := input.Type
	if msgType == "" {
		msgType = model.MessageTypeText
	}

	var msg model.Message
	err := r.db.QueryRow(ctx,
		`INSERT INTO messages (chat_id, sender_id, content, reply_to_id, type, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, chat_id, sender_id, content, reply_to_id, type, metadata, is_deleted, deleted_for_all, created_at`,
		input.ChatID, input.SenderID, input.Content, input.ReplyToID, msgType, input.Metadata,
	).Scan(
		&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content, &msg.ReplyToID,
		&msg.Type, &msg.Metadata, &msg.IsDeleted, &msg.DeletedForAll, &msg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	return &msg, nil
}

func (r *pgMessageRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Message, error) {
	var msg model.Message
	err := r.db.QueryRow(ctx,
		`SELECT id, chat_id, sender_id, content, reply_to_id, type, metadata, is_deleted, deleted_for_all, created_at
		 FROM messages WHERE id = $1`, id,
	).Scan(
		&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content, &msg.ReplyToID,
		&msg.Type, &msg.Metadata, &msg.IsDeleted, &msg.DeletedForAll, &msg.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("message", id.String())
		}
		return nil, fmt.Errorf("find message by id: %w", err)
	}

	return &msg, nil
}

func (r *pgMessageRepository) ListByChat(ctx context.Context, chatID uuid.UUID, cursor *time.Time, limit int) ([]*model.Message, error) {
	if limit <= 0 {
		limit = 50
	}

	var rows pgx.Rows
	var err error

	if cursor != nil {
		rows, err = r.db.Query(ctx,
			`SELECT id, chat_id, sender_id, content, reply_to_id, type, metadata, is_deleted, deleted_for_all, created_at
			 FROM messages
			 WHERE chat_id = $1 AND created_at < $2
			 ORDER BY created_at DESC
			 LIMIT $3`,
			chatID, *cursor, limit,
		)
	} else {
		rows, err = r.db.Query(ctx,
			`SELECT id, chat_id, sender_id, content, reply_to_id, type, metadata, is_deleted, deleted_for_all, created_at
			 FROM messages
			 WHERE chat_id = $1
			 ORDER BY created_at DESC
			 LIMIT $2`,
			chatID, limit,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list messages by chat: %w", err)
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		var msg model.Message
		if err := rows.Scan(
			&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content, &msg.ReplyToID,
			&msg.Type, &msg.Metadata, &msg.IsDeleted, &msg.DeletedForAll, &msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan message row: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate message rows: %w", err)
	}

	return messages, nil
}

func (r *pgMessageRepository) MarkAsDeleted(ctx context.Context, id uuid.UUID, forAll bool) error {
	result, err := r.db.Exec(ctx,
		`UPDATE messages SET is_deleted = true, deleted_for_all = $2 WHERE id = $1`,
		id, forAll,
	)
	if err != nil {
		return fmt.Errorf("mark message as deleted: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("message", id.String())
	}

	return nil
}

func (r *pgMessageRepository) Search(ctx context.Context, chatID uuid.UUID, query string) ([]*model.Message, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, chat_id, sender_id, content, reply_to_id, type, metadata, is_deleted, deleted_for_all, created_at
		 FROM messages
		 WHERE chat_id = $1 AND content ILIKE '%' || $2 || '%' AND is_deleted = false
		 ORDER BY created_at DESC
		 LIMIT 100`,
		chatID, query,
	)
	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		var msg model.Message
		if err := rows.Scan(
			&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content, &msg.ReplyToID,
			&msg.Type, &msg.Metadata, &msg.IsDeleted, &msg.DeletedForAll, &msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan search message row: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search message rows: %w", err)
	}

	return messages, nil
}
