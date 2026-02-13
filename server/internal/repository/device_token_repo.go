package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/otoritech/chatat/internal/model"
)

// DeviceTokenRepository defines operations for managing device tokens.
type DeviceTokenRepository interface {
	Upsert(ctx context.Context, userID uuid.UUID, token, platform string) (*model.DeviceToken, error)
	Delete(ctx context.Context, userID uuid.UUID, token string) error
	DeleteByToken(ctx context.Context, token string) error
	FindByUser(ctx context.Context, userID uuid.UUID) ([]*model.DeviceToken, error)
	FindByUsers(ctx context.Context, userIDs []uuid.UUID) ([]*model.DeviceToken, error)
	DeleteStale(ctx context.Context, days int) (int64, error)
}

type pgDeviceTokenRepository struct {
	db *pgxpool.Pool
}

// NewDeviceTokenRepository creates a new PostgreSQL-backed DeviceTokenRepository.
func NewDeviceTokenRepository(db *pgxpool.Pool) DeviceTokenRepository {
	return &pgDeviceTokenRepository{db: db}
}

func (r *pgDeviceTokenRepository) Upsert(ctx context.Context, userID uuid.UUID, token, platform string) (*model.DeviceToken, error) {
	var dt model.DeviceToken
	err := r.db.QueryRow(ctx,
		`INSERT INTO device_tokens (user_id, token, platform)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, token)
		 DO UPDATE SET platform = EXCLUDED.platform, updated_at = NOW()
		 RETURNING id, user_id, token, platform, created_at, updated_at`,
		userID, token, platform,
	).Scan(&dt.ID, &dt.UserID, &dt.Token, &dt.Platform, &dt.CreatedAt, &dt.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert device token: %w", err)
	}
	return &dt, nil
}

func (r *pgDeviceTokenRepository) Delete(ctx context.Context, userID uuid.UUID, token string) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM device_tokens WHERE user_id = $1 AND token = $2`,
		userID, token,
	)
	if err != nil {
		return fmt.Errorf("delete device token: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("device token not found")
	}
	return nil
}

func (r *pgDeviceTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM device_tokens WHERE token = $1`,
		token,
	)
	if err != nil {
		return fmt.Errorf("delete device token by token: %w", err)
	}
	return nil
}

func (r *pgDeviceTokenRepository) FindByUser(ctx context.Context, userID uuid.UUID) ([]*model.DeviceToken, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, token, platform, created_at, updated_at
		 FROM device_tokens WHERE user_id = $1
		 ORDER BY updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("find device tokens by user: %w", err)
	}
	defer rows.Close()
	return scanDeviceTokens(rows)
}

func (r *pgDeviceTokenRepository) FindByUsers(ctx context.Context, userIDs []uuid.UUID) ([]*model.DeviceToken, error) {
	if len(userIDs) == 0 {
		return []*model.DeviceToken{}, nil
	}
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, token, platform, created_at, updated_at
		 FROM device_tokens WHERE user_id = ANY($1)
		 ORDER BY user_id, updated_at DESC`,
		userIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("find device tokens by users: %w", err)
	}
	defer rows.Close()
	return scanDeviceTokens(rows)
}

func (r *pgDeviceTokenRepository) DeleteStale(ctx context.Context, days int) (int64, error) {
	result, err := r.db.Exec(ctx,
		`DELETE FROM device_tokens WHERE updated_at < NOW() - INTERVAL '1 day' * $1`,
		days,
	)
	if err != nil {
		return 0, fmt.Errorf("delete stale device tokens: %w", err)
	}
	return result.RowsAffected(), nil
}

func scanDeviceTokens(rows pgx.Rows) ([]*model.DeviceToken, error) {
	var tokens []*model.DeviceToken
	for rows.Next() {
		var dt model.DeviceToken
		if err := rows.Scan(&dt.ID, &dt.UserID, &dt.Token, &dt.Platform, &dt.CreatedAt, &dt.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan device token: %w", err)
		}
		tokens = append(tokens, &dt)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate device tokens: %w", err)
	}
	if tokens == nil {
		tokens = []*model.DeviceToken{}
	}
	return tokens, nil
}

// Ensure interface is implemented.
var _ DeviceTokenRepository = (*pgDeviceTokenRepository)(nil)

// Ensure rows type is used.
var _ = errors.New
