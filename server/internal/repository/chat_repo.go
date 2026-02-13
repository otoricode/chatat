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

// ChatRepository defines operations for managing chats and chat members.
type ChatRepository interface {
	Create(ctx context.Context, input model.CreateChatInput) (*model.Chat, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Chat, error)
	FindPersonalChat(ctx context.Context, userID1, userID2 uuid.UUID) (*model.Chat, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.ChatWithLastMessage, error)
	AddMember(ctx context.Context, chatID, userID uuid.UUID, role model.MemberRole) error
	RemoveMember(ctx context.Context, chatID, userID uuid.UUID) error
	GetMembers(ctx context.Context, chatID uuid.UUID) ([]*model.ChatMember, error)
	Update(ctx context.Context, id uuid.UUID, input model.UpdateChatInput) (*model.Chat, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Pin(ctx context.Context, id uuid.UUID) error
	Unpin(ctx context.Context, id uuid.UUID) error
}

type pgChatRepository struct {
	db *pgxpool.Pool
}

// NewChatRepository creates a new PostgreSQL-backed ChatRepository.
func NewChatRepository(db *pgxpool.Pool) ChatRepository {
	return &pgChatRepository{db: db}
}

func (r *pgChatRepository) Create(ctx context.Context, input model.CreateChatInput) (*model.Chat, error) {
	var chat model.Chat
	err := r.db.QueryRow(ctx,
		`INSERT INTO chats (type, name, icon, description, created_by)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, type, name, icon, description, created_by, pinned_at, created_at, updated_at`,
		input.Type, input.Name, input.Icon, input.Description, input.CreatedBy,
	).Scan(
		&chat.ID, &chat.Type, &chat.Name, &chat.Icon, &chat.Description,
		&chat.CreatedBy, &chat.PinnedAt, &chat.CreatedAt, &chat.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}

	return &chat, nil
}

func (r *pgChatRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Chat, error) {
	var chat model.Chat
	err := r.db.QueryRow(ctx,
		`SELECT id, type, name, icon, description, created_by, pinned_at, created_at, updated_at
		 FROM chats WHERE id = $1`, id,
	).Scan(
		&chat.ID, &chat.Type, &chat.Name, &chat.Icon, &chat.Description,
		&chat.CreatedBy, &chat.PinnedAt, &chat.CreatedAt, &chat.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("chat", id.String())
		}
		return nil, fmt.Errorf("find chat by id: %w", err)
	}

	return &chat, nil
}

func (r *pgChatRepository) FindPersonalChat(ctx context.Context, userID1, userID2 uuid.UUID) (*model.Chat, error) {
	var chat model.Chat
	err := r.db.QueryRow(ctx,
		`SELECT c.id, c.type, c.name, c.icon, c.description, c.created_by, c.pinned_at, c.created_at, c.updated_at
		 FROM chats c
		 JOIN chat_members cm1 ON c.id = cm1.chat_id AND cm1.user_id = $1
		 JOIN chat_members cm2 ON c.id = cm2.chat_id AND cm2.user_id = $2
		 WHERE c.type = 'personal'`, userID1, userID2,
	).Scan(
		&chat.ID, &chat.Type, &chat.Name, &chat.Icon, &chat.Description,
		&chat.CreatedBy, &chat.PinnedAt, &chat.CreatedAt, &chat.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("personal chat", userID1.String()+":"+userID2.String())
		}
		return nil, fmt.Errorf("find personal chat: %w", err)
	}

	return &chat, nil
}

func (r *pgChatRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.ChatWithLastMessage, error) {
	rows, err := r.db.Query(ctx,
		`SELECT c.id, c.type, c.name, c.icon, c.description, c.created_by, c.pinned_at, c.created_at, c.updated_at,
		        (SELECT m.content FROM messages m WHERE m.chat_id = c.id ORDER BY m.created_at DESC LIMIT 1) as last_message
		 FROM chats c
		 JOIN chat_members cm ON c.id = cm.chat_id
		 WHERE cm.user_id = $1
		 ORDER BY c.updated_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list chats by user: %w", err)
	}
	defer rows.Close()

	var chats []*model.ChatWithLastMessage
	for rows.Next() {
		var cwm model.ChatWithLastMessage
		if err := rows.Scan(
			&cwm.Chat.ID, &cwm.Chat.Type, &cwm.Chat.Name, &cwm.Chat.Icon,
			&cwm.Chat.Description, &cwm.Chat.CreatedBy, &cwm.Chat.PinnedAt,
			&cwm.Chat.CreatedAt, &cwm.Chat.UpdatedAt, &cwm.LastMessage,
		); err != nil {
			return nil, fmt.Errorf("scan chat row: %w", err)
		}
		chats = append(chats, &cwm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat rows: %w", err)
	}

	return chats, nil
}

func (r *pgChatRepository) AddMember(ctx context.Context, chatID, userID uuid.UUID, role model.MemberRole) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO chat_members (chat_id, user_id, role) VALUES ($1, $2, $3)
		 ON CONFLICT (chat_id, user_id) DO NOTHING`,
		chatID, userID, role,
	)
	if err != nil {
		return fmt.Errorf("add chat member: %w", err)
	}

	return nil
}

func (r *pgChatRepository) RemoveMember(ctx context.Context, chatID, userID uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM chat_members WHERE chat_id = $1 AND user_id = $2`,
		chatID, userID,
	)
	if err != nil {
		return fmt.Errorf("remove chat member: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("chat member", userID.String())
	}

	return nil
}

func (r *pgChatRepository) GetMembers(ctx context.Context, chatID uuid.UUID) ([]*model.ChatMember, error) {
	rows, err := r.db.Query(ctx,
		`SELECT chat_id, user_id, role, joined_at
		 FROM chat_members WHERE chat_id = $1
		 ORDER BY joined_at`, chatID,
	)
	if err != nil {
		return nil, fmt.Errorf("get chat members: %w", err)
	}
	defer rows.Close()

	var members []*model.ChatMember
	for rows.Next() {
		var m model.ChatMember
		if err := rows.Scan(&m.ChatID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan chat member: %w", err)
		}
		members = append(members, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat member rows: %w", err)
	}

	return members, nil
}

func (r *pgChatRepository) Update(ctx context.Context, id uuid.UUID, input model.UpdateChatInput) (*model.Chat, error) {
	var chat model.Chat
	err := r.db.QueryRow(ctx,
		`UPDATE chats SET
		   name = COALESCE($2, name),
		   icon = COALESCE($3, icon),
		   description = COALESCE($4, description),
		   updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, type, name, icon, description, created_by, pinned_at, created_at, updated_at`,
		id, input.Name, input.Icon, input.Description,
	).Scan(
		&chat.ID, &chat.Type, &chat.Name, &chat.Icon, &chat.Description,
		&chat.CreatedBy, &chat.PinnedAt, &chat.CreatedAt, &chat.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("chat", id.String())
		}
		return nil, fmt.Errorf("update chat: %w", err)
	}

	return &chat, nil
}

func (r *pgChatRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM chats WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("delete chat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("chat", id.String())
	}

	return nil
}

func (r *pgChatRepository) Pin(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`UPDATE chats SET pinned_at = NOW() WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("pin chat: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperror.NotFound("chat", id.String())
	}
	return nil
}

func (r *pgChatRepository) Unpin(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`UPDATE chats SET pinned_at = NULL WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("unpin chat: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperror.NotFound("chat", id.String())
	}
	return nil
}
