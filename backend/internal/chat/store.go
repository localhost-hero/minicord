// Package chat persists room text messages and broadcasts them to connected
// participants in real time.
package chat

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const maxMessageLength = 4000

// Message is a single persisted chat message.
type Message struct {
	ID        int64     `json:"id"`
	Room      string    `json:"room"`
	Identity  string    `json:"identity"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

// Store persists chat messages in the database.
type Store struct {
	db *sql.DB
}

// NewStore returns a chat message store backed by db.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// SaveMessage validates and persists a chat message, returning the stored record.
func (s *Store) SaveMessage(room, identity, body string) (Message, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return Message{}, fmt.Errorf("chat: message body must not be empty")
	}
	if len(body) > maxMessageLength {
		return Message{}, fmt.Errorf("chat: message body must be at most %d characters", maxMessageLength)
	}

	now := time.Now().UTC()
	result, err := s.db.Exec(
		`INSERT INTO messages (room, identity, body, created_at) VALUES (?, ?, ?, ?)`,
		room, identity, body, now.Unix(),
	)
	if err != nil {
		return Message{}, fmt.Errorf("chat: insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Message{}, fmt.Errorf("chat: read inserted id: %w", err)
	}

	return Message{ID: id, Room: room, Identity: identity, Body: body, CreatedAt: now}, nil
}

// ListMessages returns up to limit of the most recent messages for a room, oldest first.
func (s *Store) ListMessages(room string, limit int) ([]Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := s.db.Query(
		`SELECT id, room, identity, body, created_at FROM messages
		 WHERE room = ? ORDER BY id DESC LIMIT ?`,
		room, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("chat: query: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var (
			message   Message
			createdAt int64
		)
		if err := rows.Scan(&message.ID, &message.Room, &message.Identity, &message.Body, &createdAt); err != nil {
			return nil, fmt.Errorf("chat: scan: %w", err)
		}
		message.CreatedAt = time.Unix(createdAt, 0).UTC()
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chat: rows: %w", err)
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	if messages == nil {
		messages = []Message{}
	}

	return messages, nil
}
