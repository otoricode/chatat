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

// TopicRepository defines operations for managing topics.
type TopicRepository interface {
	Create(ctx context.Context, input model.CreateTopicInput) (*model.Topic, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Topic, error)
	ListByParent(ctx context.Context, parentID uuid.UUID) ([]*model.Topic, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.Topic, error)
	AddMember(ctx context.Context, topicID, userID uuid.UUID, role model.MemberRole) error
	RemoveMember(ctx context.Context, topicID, userID uuid.UUID) error
	GetMembers(ctx context.Context, topicID uuid.UUID) ([]*model.TopicMember, error)
	Update(ctx context.Context, id uuid.UUID, input model.UpdateTopicInput) (*model.Topic, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type pgTopicRepository struct {
	db *pgxpool.Pool
}

// NewTopicRepository creates a new PostgreSQL-backed TopicRepository.
func NewTopicRepository(db *pgxpool.Pool) TopicRepository {
	return &pgTopicRepository{db: db}
}

func (r *pgTopicRepository) Create(ctx context.Context, input model.CreateTopicInput) (*model.Topic, error) {
	icon := input.Icon
	if icon == "" {
		icon = "💬"
	}

	var topic model.Topic
	err := r.db.QueryRow(ctx,
		`INSERT INTO topics (name, icon, description, parent_type, parent_id, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, name, icon, description, parent_type, parent_id, created_by, created_at, updated_at`,
		input.Name, icon, input.Description, input.ParentType, input.ParentID, input.CreatedBy,
	).Scan(
		&topic.ID, &topic.Name, &topic.Icon, &topic.Description,
		&topic.ParentType, &topic.ParentID, &topic.CreatedBy, &topic.CreatedAt, &topic.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create topic: %w", err)
	}

	return &topic, nil
}

func (r *pgTopicRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Topic, error) {
	var topic model.Topic
	err := r.db.QueryRow(ctx,
		`SELECT id, name, icon, description, parent_type, parent_id, created_by, created_at, updated_at
		 FROM topics WHERE id = $1`, id,
	).Scan(
		&topic.ID, &topic.Name, &topic.Icon, &topic.Description,
		&topic.ParentType, &topic.ParentID, &topic.CreatedBy, &topic.CreatedAt, &topic.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("topic", id.String())
		}
		return nil, fmt.Errorf("find topic by id: %w", err)
	}

	return &topic, nil
}

func (r *pgTopicRepository) ListByParent(ctx context.Context, parentID uuid.UUID) ([]*model.Topic, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, icon, description, parent_type, parent_id, created_by, created_at, updated_at
		 FROM topics WHERE parent_id = $1
		 ORDER BY created_at`, parentID,
	)
	if err != nil {
		return nil, fmt.Errorf("list topics by parent: %w", err)
	}
	defer rows.Close()

	var topics []*model.Topic
	for rows.Next() {
		var t model.Topic
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Icon, &t.Description,
			&t.ParentType, &t.ParentID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan topic row: %w", err)
		}
		topics = append(topics, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate topic rows: %w", err)
	}

	return topics, nil
}

func (r *pgTopicRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.Topic, error) {
	rows, err := r.db.Query(ctx,
		`SELECT t.id, t.name, t.icon, t.description, t.parent_type, t.parent_id, t.created_by, t.created_at, t.updated_at
		 FROM topics t
		 JOIN topic_members tm ON t.id = tm.topic_id
		 WHERE tm.user_id = $1
		 ORDER BY t.updated_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list topics by user: %w", err)
	}
	defer rows.Close()

	var topics []*model.Topic
	for rows.Next() {
		var t model.Topic
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Icon, &t.Description,
			&t.ParentType, &t.ParentID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan topic row: %w", err)
		}
		topics = append(topics, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate topic rows: %w", err)
	}

	return topics, nil
}

func (r *pgTopicRepository) AddMember(ctx context.Context, topicID, userID uuid.UUID, role model.MemberRole) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO topic_members (topic_id, user_id, role) VALUES ($1, $2, $3)
		 ON CONFLICT (topic_id, user_id) DO NOTHING`,
		topicID, userID, role,
	)
	if err != nil {
		return fmt.Errorf("add topic member: %w", err)
	}

	return nil
}

func (r *pgTopicRepository) RemoveMember(ctx context.Context, topicID, userID uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM topic_members WHERE topic_id = $1 AND user_id = $2`,
		topicID, userID,
	)
	if err != nil {
		return fmt.Errorf("remove topic member: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("topic member", userID.String())
	}

	return nil
}

func (r *pgTopicRepository) GetMembers(ctx context.Context, topicID uuid.UUID) ([]*model.TopicMember, error) {
	rows, err := r.db.Query(ctx,
		`SELECT topic_id, user_id, role, joined_at
		 FROM topic_members WHERE topic_id = $1
		 ORDER BY joined_at`, topicID,
	)
	if err != nil {
		return nil, fmt.Errorf("get topic members: %w", err)
	}
	defer rows.Close()

	var members []*model.TopicMember
	for rows.Next() {
		var m model.TopicMember
		if err := rows.Scan(&m.TopicID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan topic member: %w", err)
		}
		members = append(members, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate topic member rows: %w", err)
	}

	return members, nil
}

func (r *pgTopicRepository) Update(ctx context.Context, id uuid.UUID, input model.UpdateTopicInput) (*model.Topic, error) {
	var topic model.Topic
	err := r.db.QueryRow(ctx,
		`UPDATE topics SET
		   name = COALESCE($2, name),
		   icon = COALESCE($3, icon),
		   description = COALESCE($4, description),
		   updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, name, icon, description, parent_type, parent_id, created_by, created_at, updated_at`,
		id, input.Name, input.Icon, input.Description,
	).Scan(
		&topic.ID, &topic.Name, &topic.Icon, &topic.Description,
		&topic.ParentType, &topic.ParentID, &topic.CreatedBy, &topic.CreatedAt, &topic.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("topic", id.String())
		}
		return nil, fmt.Errorf("update topic: %w", err)
	}

	return &topic, nil
}

func (r *pgTopicRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM topics WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("delete topic: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("topic", id.String())
	}

	return nil
}
