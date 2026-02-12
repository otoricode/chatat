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

// TopicMessageRepository defines operations for managing topic messages.
type TopicMessageRepository interface {
	Create(ctx context.Context, input model.CreateTopicMessageInput) (*model.TopicMessage, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.TopicMessage, error)
	ListByTopic(ctx context.Context, topicID uuid.UUID, cursor *time.Time, limit int) ([]*model.TopicMessage, error)
	MarkAsDeleted(ctx context.Context, id uuid.UUID, forAll bool) error
}

type pgTopicMessageRepository struct {
	db *pgxpool.Pool
}

// NewTopicMessageRepository creates a new PostgreSQL-backed TopicMessageRepository.
func NewTopicMessageRepository(db *pgxpool.Pool) TopicMessageRepository {
	return &pgTopicMessageRepository{db: db}
}

func (r *pgTopicMessageRepository) Create(ctx context.Context, input model.CreateTopicMessageInput) (*model.TopicMessage, error) {
	msgType := input.Type
	if msgType == "" {
		msgType = model.MessageTypeText
	}

	var msg model.TopicMessage
	err := r.db.QueryRow(ctx,
		`INSERT INTO topic_messages (topic_id, sender_id, content, reply_to_id, type)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, topic_id, sender_id, content, reply_to_id, type, is_deleted, deleted_for_all, created_at`,
		input.TopicID, input.SenderID, input.Content, input.ReplyToID, msgType,
	).Scan(
		&msg.ID, &msg.TopicID, &msg.SenderID, &msg.Content, &msg.ReplyToID,
		&msg.Type, &msg.IsDeleted, &msg.DeletedForAll, &msg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create topic message: %w", err)
	}

	return &msg, nil
}

func (r *pgTopicMessageRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.TopicMessage, error) {
	var msg model.TopicMessage
	err := r.db.QueryRow(ctx,
		`SELECT id, topic_id, sender_id, content, reply_to_id, type, is_deleted, deleted_for_all, created_at
		 FROM topic_messages WHERE id = $1`, id,
	).Scan(
		&msg.ID, &msg.TopicID, &msg.SenderID, &msg.Content, &msg.ReplyToID,
		&msg.Type, &msg.IsDeleted, &msg.DeletedForAll, &msg.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("topic message", id.String())
		}
		return nil, fmt.Errorf("find topic message by id: %w", err)
	}

	return &msg, nil
}

func (r *pgTopicMessageRepository) ListByTopic(ctx context.Context, topicID uuid.UUID, cursor *time.Time, limit int) ([]*model.TopicMessage, error) {
	if limit <= 0 {
		limit = 50
	}

	var rows pgx.Rows
	var err error

	if cursor != nil {
		rows, err = r.db.Query(ctx,
			`SELECT id, topic_id, sender_id, content, reply_to_id, type, is_deleted, deleted_for_all, created_at
			 FROM topic_messages
			 WHERE topic_id = $1 AND created_at < $2
			 ORDER BY created_at DESC
			 LIMIT $3`,
			topicID, *cursor, limit,
		)
	} else {
		rows, err = r.db.Query(ctx,
			`SELECT id, topic_id, sender_id, content, reply_to_id, type, is_deleted, deleted_for_all, created_at
			 FROM topic_messages
			 WHERE topic_id = $1
			 ORDER BY created_at DESC
			 LIMIT $2`,
			topicID, limit,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list topic messages: %w", err)
	}
	defer rows.Close()

	var messages []*model.TopicMessage
	for rows.Next() {
		var msg model.TopicMessage
		if err := rows.Scan(
			&msg.ID, &msg.TopicID, &msg.SenderID, &msg.Content, &msg.ReplyToID,
			&msg.Type, &msg.IsDeleted, &msg.DeletedForAll, &msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan topic message row: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate topic message rows: %w", err)
	}

	return messages, nil
}

func (r *pgTopicMessageRepository) MarkAsDeleted(ctx context.Context, id uuid.UUID, forAll bool) error {
	result, err := r.db.Exec(ctx,
		`UPDATE topic_messages SET is_deleted = true, deleted_for_all = $2 WHERE id = $1`,
		id, forAll,
	)
	if err != nil {
		return fmt.Errorf("mark topic message as deleted: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("topic message", id.String())
	}

	return nil
}
