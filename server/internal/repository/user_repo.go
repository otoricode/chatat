// Package repository provides data access interfaces and their PostgreSQL implementations.
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

// UserRepository defines operations for managing users.
type UserRepository interface {
	Create(ctx context.Context, input model.CreateUserInput) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByPhone(ctx context.Context, phone string) (*model.User, error)
	FindByPhones(ctx context.Context, phones []string) ([]*model.User, error)
	Update(ctx context.Context, id uuid.UUID, input model.UpdateUserInput) (*model.User, error)
	UpdateLastSeen(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type pgUserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new PostgreSQL-backed UserRepository.
func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &pgUserRepository{db: db}
}

func (r *pgUserRepository) Create(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
	avatar := input.Avatar
	if avatar == "" {
		avatar = "\U0001F464"
	}

	var user model.User
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (phone, name, avatar)
		 VALUES ($1, $2, $3)
		 RETURNING id, phone, name, avatar, status, last_seen, created_at, updated_at`,
		input.Phone, input.Name, avatar,
	).Scan(
		&user.ID, &user.Phone, &user.Name, &user.Avatar,
		&user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, apperror.Conflict("phone number already registered")
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

func (r *pgUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx,
		`SELECT id, phone, name, avatar, status, last_seen, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(
		&user.ID, &user.Phone, &user.Name, &user.Avatar,
		&user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("user", id.String())
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return &user, nil
}

func (r *pgUserRepository) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx,
		`SELECT id, phone, name, avatar, status, last_seen, created_at, updated_at
		 FROM users WHERE phone = $1`, phone,
	).Scan(
		&user.ID, &user.Phone, &user.Name, &user.Avatar,
		&user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("user", phone)
		}
		return nil, fmt.Errorf("find user by phone: %w", err)
	}

	return &user, nil
}

func (r *pgUserRepository) FindByPhones(ctx context.Context, phones []string) ([]*model.User, error) {
	if len(phones) == 0 {
		return []*model.User{}, nil
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, phone, name, avatar, status, last_seen, created_at, updated_at
		 FROM users WHERE phone = ANY($1)`, phones,
	)
	if err != nil {
		return nil, fmt.Errorf("find users by phones: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(
			&user.ID, &user.Phone, &user.Name, &user.Avatar,
			&user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user row: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user rows: %w", err)
	}

	return users, nil
}

func (r *pgUserRepository) Update(ctx context.Context, id uuid.UUID, input model.UpdateUserInput) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx,
		`UPDATE users SET
		   name = COALESCE($2, name),
		   avatar = COALESCE($3, avatar),
		   status = COALESCE($4, status),
		   updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, phone, name, avatar, status, last_seen, created_at, updated_at`,
		id, input.Name, input.Avatar, input.Status,
	).Scan(
		&user.ID, &user.Phone, &user.Name, &user.Avatar,
		&user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("user", id.String())
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &user, nil
}

func (r *pgUserRepository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`UPDATE users SET last_seen = $2 WHERE id = $1`,
		id, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update last seen: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("user", id.String())
	}

	return nil
}

func (r *pgUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM users WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFound("user", id.String())
	}

	return nil
}
