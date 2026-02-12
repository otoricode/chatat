package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserContact represents a cached contact relationship.
type UserContact struct {
	UserID        uuid.UUID `json:"userId"`
	ContactUserID uuid.UUID `json:"contactUserId"`
	ContactName   string    `json:"contactName"`
	SyncedAt      time.Time `json:"syncedAt"`
}

// ContactRepository defines operations for managing user contacts.
type ContactRepository interface {
	Upsert(ctx context.Context, userID, contactUserID uuid.UUID, contactName string) error
	UpsertBatch(ctx context.Context, userID uuid.UUID, contacts []ContactUpsertInput) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]UserContact, error)
	FindContactsOf(ctx context.Context, contactUserID uuid.UUID) ([]uuid.UUID, error)
	Delete(ctx context.Context, userID, contactUserID uuid.UUID) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}

// ContactUpsertInput holds data for upserting a single contact.
type ContactUpsertInput struct {
	ContactUserID uuid.UUID
	ContactName   string
}

type pgContactRepository struct {
	db *pgxpool.Pool
}

// NewContactRepository creates a new PostgreSQL-backed ContactRepository.
func NewContactRepository(db *pgxpool.Pool) ContactRepository {
	return &pgContactRepository{db: db}
}

func (r *pgContactRepository) Upsert(ctx context.Context, userID, contactUserID uuid.UUID, contactName string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_contacts (user_id, contact_user_id, contact_name, synced_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (user_id, contact_user_id) DO UPDATE SET
		   contact_name = EXCLUDED.contact_name,
		   synced_at = NOW()`,
		userID, contactUserID, contactName,
	)
	if err != nil {
		return fmt.Errorf("upsert contact: %w", err)
	}
	return nil
}

func (r *pgContactRepository) UpsertBatch(ctx context.Context, userID uuid.UUID, contacts []ContactUpsertInput) error {
	if len(contacts) == 0 {
		return nil
	}

	batch := &pgxpool.Pool{}
	_ = batch // unused, we use a transaction instead

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, c := range contacts {
		_, err := tx.Exec(ctx,
			`INSERT INTO user_contacts (user_id, contact_user_id, contact_name, synced_at)
			 VALUES ($1, $2, $3, NOW())
			 ON CONFLICT (user_id, contact_user_id) DO UPDATE SET
			   contact_name = EXCLUDED.contact_name,
			   synced_at = NOW()`,
			userID, c.ContactUserID, c.ContactName,
		)
		if err != nil {
			return fmt.Errorf("upsert contact %s: %w", c.ContactUserID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *pgContactRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]UserContact, error) {
	rows, err := r.db.Query(ctx,
		`SELECT user_id, contact_user_id, COALESCE(contact_name, ''), synced_at
		 FROM user_contacts
		 WHERE user_id = $1
		 ORDER BY contact_name ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("find contacts: %w", err)
	}
	defer rows.Close()

	var contacts []UserContact
	for rows.Next() {
		var c UserContact
		if err := rows.Scan(&c.UserID, &c.ContactUserID, &c.ContactName, &c.SyncedAt); err != nil {
			return nil, fmt.Errorf("scan contact: %w", err)
		}
		contacts = append(contacts, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate contacts: %w", err)
	}

	return contacts, nil
}

func (r *pgContactRepository) FindContactsOf(ctx context.Context, contactUserID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx,
		`SELECT user_id FROM user_contacts WHERE contact_user_id = $1`,
		contactUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("find contacts of: %w", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan contact of: %w", err)
		}
		userIDs = append(userIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate contacts of: %w", err)
	}

	return userIDs, nil
}

func (r *pgContactRepository) Delete(ctx context.Context, userID, contactUserID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM user_contacts WHERE user_id = $1 AND contact_user_id = $2`,
		userID, contactUserID,
	)
	if err != nil {
		return fmt.Errorf("delete contact: %w", err)
	}
	return nil
}

func (r *pgContactRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM user_contacts WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("delete all contacts: %w", err)
	}
	return nil
}
