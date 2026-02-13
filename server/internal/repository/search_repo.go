package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/otoritech/chatat/internal/model"
)

// SearchRepository defines operations for full-text search across the database.
type SearchRepository interface {
	SearchMessages(ctx context.Context, userID uuid.UUID, query string, offset, limit int) ([]*model.MessageSearchRow, error)
	SearchMessagesInChat(ctx context.Context, chatID uuid.UUID, query string, offset, limit int) ([]*model.MessageSearchRow, error)
	SearchDocuments(ctx context.Context, userID uuid.UUID, query string, offset, limit int) ([]*model.DocumentSearchRow, error)
	SearchContacts(ctx context.Context, userID uuid.UUID, query string, limit int) ([]*model.User, error)
	SearchEntities(ctx context.Context, userID uuid.UUID, query string, limit int) ([]*model.Entity, error)
}

type pgSearchRepository struct {
	db *pgxpool.Pool
}

// NewSearchRepository creates a new PostgreSQL-backed SearchRepository.
func NewSearchRepository(db *pgxpool.Pool) SearchRepository {
	return &pgSearchRepository{db: db}
}

// BuildTSQuery converts user input into a tsquery with prefix matching.
func BuildTSQuery(query string) string {
	words := strings.Fields(strings.TrimSpace(query))
	if len(words) == 0 {
		return ""
	}
	escaped := make([]string, len(words))
	for i, w := range words {
		// Escape special characters for tsquery
		w = strings.ReplaceAll(w, "'", "''")
		w = strings.ReplaceAll(w, "\\", "\\\\")
		escaped[i] = "'" + w + "':*"
	}
	return strings.Join(escaped, " & ")
}

func (r *pgSearchRepository) SearchMessages(ctx context.Context, userID uuid.UUID, query string, offset, limit int) ([]*model.MessageSearchRow, error) {
	tsq := BuildTSQuery(query)
	if tsq == "" {
		return []*model.MessageSearchRow{}, nil
	}

	rows, err := r.db.Query(ctx,
		`SELECT m.id, m.chat_id, m.sender_id, m.content, m.type, m.created_at,
		        COALESCE(c.name, '') AS chat_name,
		        COALESCE(u.name, '') AS sender_name,
		        ts_headline('indonesian', m.content,
		            to_tsquery('indonesian', $1),
		            'MaxWords=30, MinWords=10, StartSel=<mark>, StopSel=</mark>'
		        ) AS highlight
		 FROM messages m
		 JOIN chat_members cm ON cm.chat_id = m.chat_id AND cm.user_id = $2
		 JOIN chats c ON c.id = m.chat_id
		 JOIN users u ON u.id = m.sender_id
		 WHERE m.search_vector @@ to_tsquery('indonesian', $1)
		   AND m.is_deleted = false
		 ORDER BY m.created_at DESC
		 OFFSET $3 LIMIT $4`,
		tsq, userID, offset, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}
	defer rows.Close()

	var results []*model.MessageSearchRow
	for rows.Next() {
		var r model.MessageSearchRow
		if err := rows.Scan(
			&r.ID, &r.ChatID, &r.SenderID, &r.Content, &r.Type, &r.CreatedAt,
			&r.ChatName, &r.SenderName, &r.Highlight,
		); err != nil {
			return nil, fmt.Errorf("scan message search row: %w", err)
		}
		results = append(results, &r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate message search rows: %w", err)
	}
	if results == nil {
		results = []*model.MessageSearchRow{}
	}
	return results, nil
}

func (r *pgSearchRepository) SearchMessagesInChat(ctx context.Context, chatID uuid.UUID, query string, offset, limit int) ([]*model.MessageSearchRow, error) {
	tsq := BuildTSQuery(query)
	if tsq == "" {
		return []*model.MessageSearchRow{}, nil
	}

	rows, err := r.db.Query(ctx,
		`SELECT m.id, m.chat_id, m.sender_id, m.content, m.type, m.created_at,
		        COALESCE(c.name, '') AS chat_name,
		        COALESCE(u.name, '') AS sender_name,
		        ts_headline('indonesian', m.content,
		            to_tsquery('indonesian', $1),
		            'MaxWords=30, MinWords=10, StartSel=<mark>, StopSel=</mark>'
		        ) AS highlight
		 FROM messages m
		 JOIN chats c ON c.id = m.chat_id
		 JOIN users u ON u.id = m.sender_id
		 WHERE m.chat_id = $2
		   AND m.search_vector @@ to_tsquery('indonesian', $1)
		   AND m.is_deleted = false
		 ORDER BY m.created_at DESC
		 OFFSET $3 LIMIT $4`,
		tsq, chatID, offset, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search messages in chat: %w", err)
	}
	defer rows.Close()

	var results []*model.MessageSearchRow
	for rows.Next() {
		var r model.MessageSearchRow
		if err := rows.Scan(
			&r.ID, &r.ChatID, &r.SenderID, &r.Content, &r.Type, &r.CreatedAt,
			&r.ChatName, &r.SenderName, &r.Highlight,
		); err != nil {
			return nil, fmt.Errorf("scan chat message search row: %w", err)
		}
		results = append(results, &r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat message search rows: %w", err)
	}
	if results == nil {
		results = []*model.MessageSearchRow{}
	}
	return results, nil
}

func (r *pgSearchRepository) SearchDocuments(ctx context.Context, userID uuid.UUID, query string, offset, limit int) ([]*model.DocumentSearchRow, error) {
	tsq := BuildTSQuery(query)
	if tsq == "" {
		return []*model.DocumentSearchRow{}, nil
	}

	rows, err := r.db.Query(ctx,
		`SELECT DISTINCT ON (d.id) d.id, d.title, d.icon, d.owner_id, d.locked, d.updated_at,
		        ts_headline('indonesian', COALESCE(b.content, d.title),
		            to_tsquery('indonesian', $1),
		            'MaxWords=30, MinWords=10, StartSel=<mark>, StopSel=</mark>'
		        ) AS highlight
		 FROM documents d
		 LEFT JOIN blocks b ON b.document_id = d.id
		   AND b.search_vector @@ to_tsquery('indonesian', $1)
		 WHERE (
		   d.search_vector @@ to_tsquery('indonesian', $1)
		   OR b.search_vector @@ to_tsquery('indonesian', $1)
		 )
		 AND (
		   d.owner_id = $2
		   OR d.id IN (SELECT document_id FROM document_collaborators WHERE user_id = $2)
		   OR d.chat_id IN (SELECT chat_id FROM chat_members WHERE user_id = $2)
		   OR d.topic_id IN (
		     SELECT t.id FROM topics t
		     JOIN chat_members cm ON cm.chat_id = t.chat_id AND cm.user_id = $2
		   )
		 )
		 ORDER BY d.id, d.updated_at DESC
		 OFFSET $3 LIMIT $4`,
		tsq, userID, offset, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search documents: %w", err)
	}
	defer rows.Close()

	var results []*model.DocumentSearchRow
	for rows.Next() {
		var r model.DocumentSearchRow
		if err := rows.Scan(
			&r.ID, &r.Title, &r.Icon, &r.OwnerID, &r.Locked, &r.UpdatedAt, &r.Highlight,
		); err != nil {
			return nil, fmt.Errorf("scan document search row: %w", err)
		}
		results = append(results, &r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate document search rows: %w", err)
	}
	if results == nil {
		results = []*model.DocumentSearchRow{}
	}
	return results, nil
}

func (r *pgSearchRepository) SearchContacts(ctx context.Context, userID uuid.UUID, query string, limit int) ([]*model.User, error) {
	tsq := BuildTSQuery(query)
	if tsq == "" {
		return []*model.User{}, nil
	}

	rows, err := r.db.Query(ctx,
		`SELECT u.id, u.phone, u.name, u.avatar, u.status, u.last_seen, u.created_at, u.updated_at
		 FROM users u
		 JOIN user_contacts c ON c.contact_user_id = u.id AND c.user_id = $2
		 WHERE u.search_vector @@ to_tsquery('simple', $1)
		 ORDER BY u.name
		 LIMIT $3`,
		tsq, userID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search contacts: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.ID, &u.Phone, &u.Name, &u.Avatar, &u.Status, &u.LastSeen, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan contact search row: %w", err)
		}
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate contact search rows: %w", err)
	}
	if users == nil {
		users = []*model.User{}
	}
	return users, nil
}

func (r *pgSearchRepository) SearchEntities(ctx context.Context, userID uuid.UUID, query string, limit int) ([]*model.Entity, error) {
	tsq := BuildTSQuery(query)
	if tsq == "" {
		return []*model.Entity{}, nil
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, name, type, COALESCE(fields, '{}'::jsonb), owner_id, contact_user_id, created_at, updated_at
		 FROM entities
		 WHERE owner_id = $2
		   AND search_vector @@ to_tsquery('simple', $1)
		 ORDER BY name
		 LIMIT $3`,
		tsq, userID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search entities: %w", err)
	}
	defer rows.Close()

	var entities []*model.Entity
	for rows.Next() {
		var e model.Entity
		var fieldsRaw []byte
		if err := rows.Scan(
			&e.ID, &e.Name, &e.Type, &fieldsRaw, &e.OwnerID, &e.ContactUserID, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan entity search row: %w", err)
		}
		_ = json.Unmarshal(fieldsRaw, &e.Fields)
		if e.Fields == nil {
			e.Fields = make(map[string]string)
		}
		entities = append(entities, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity search rows: %w", err)
	}
	if entities == nil {
		entities = []*model.Entity{}
	}
	return entities, nil
}

// Ensure interface is implemented.
var _ SearchRepository = (*pgSearchRepository)(nil)
