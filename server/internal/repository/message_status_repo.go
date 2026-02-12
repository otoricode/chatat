package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/otoritech/chatat/internal/model"
)

// MessageStatusRepository defines operations for managing message delivery/read status.
type MessageStatusRepository interface {
	Create(ctx context.Context, messageID, userID uuid.UUID, status model.DeliveryStatus) error
	UpdateStatus(ctx context.Context, messageID, userID uuid.UUID, status model.DeliveryStatus) error
	GetStatus(ctx context.Context, messageID uuid.UUID) ([]*model.MessageStatus, error)
	MarkChatAsRead(ctx context.Context, chatID, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, chatID, userID uuid.UUID) (int, error)
}

type pgMessageStatusRepository struct {
	db *pgxpool.Pool
}

// NewMessageStatusRepository creates a new PostgreSQL-backed MessageStatusRepository.
func NewMessageStatusRepository(db *pgxpool.Pool) MessageStatusRepository {
	return &pgMessageStatusRepository{db: db}
}

func (r *pgMessageStatusRepository) Create(ctx context.Context, messageID, userID uuid.UUID, status model.DeliveryStatus) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO message_status (message_id, user_id, status) VALUES ($1, $2, $3)
		 ON CONFLICT (message_id, user_id) DO NOTHING`,
		messageID, userID, status,
	)
	if err != nil {
		return fmt.Errorf("create message status: %w", err)
	}

	return nil
}

func (r *pgMessageStatusRepository) UpdateStatus(ctx context.Context, messageID, userID uuid.UUID, status model.DeliveryStatus) error {
	_, err := r.db.Exec(ctx,
		`UPDATE message_status SET status = $3, updated_at = NOW()
		 WHERE message_id = $1 AND user_id = $2`,
		messageID, userID, status,
	)
	if err != nil {
		return fmt.Errorf("update message status: %w", err)
	}

	return nil
}

func (r *pgMessageStatusRepository) GetStatus(ctx context.Context, messageID uuid.UUID) ([]*model.MessageStatus, error) {
	rows, err := r.db.Query(ctx,
		`SELECT message_id, user_id, status, updated_at
		 FROM message_status WHERE message_id = $1`, messageID,
	)
	if err != nil {
		return nil, fmt.Errorf("get message status: %w", err)
	}
	defer rows.Close()

	var statuses []*model.MessageStatus
	for rows.Next() {
		var s model.MessageStatus
		if err := rows.Scan(&s.MessageID, &s.UserID, &s.Status, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan message status: %w", err)
		}
		statuses = append(statuses, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate message status rows: %w", err)
	}

	return statuses, nil
}

func (r *pgMessageStatusRepository) MarkChatAsRead(ctx context.Context, chatID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE message_status ms SET status = 'read', updated_at = NOW()
		 FROM messages m
		 WHERE ms.message_id = m.id AND m.chat_id = $1 AND ms.user_id = $2 AND ms.status != 'read'`,
		chatID, userID,
	)
	if err != nil {
		return fmt.Errorf("mark chat as read: %w", err)
	}

	return nil
}

func (r *pgMessageStatusRepository) GetUnreadCount(ctx context.Context, chatID, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*)
		 FROM messages m
		 LEFT JOIN message_status ms ON m.id = ms.message_id AND ms.user_id = $2
		 WHERE m.chat_id = $1 AND m.sender_id != $2
		   AND (ms.status IS NULL OR ms.status != 'read')
		   AND m.is_deleted = false`,
		chatID, userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get unread count: %w", err)
	}

	return count, nil
}
