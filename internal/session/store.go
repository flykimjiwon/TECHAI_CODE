// Author: Kim Jiwon (github.com/flykimjiwon) — forked from hanimo-code
// Package session provides SQLite-backed persistence for chat sessions,
// so users can restore prior conversations across process restarts.
package session

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	openai "github.com/sashabaranov/go-openai"

	// Pure-Go SQLite driver (no CGO required, keeps cross-compile simple).
	_ "modernc.org/sqlite"
)

// ErrNotFound is returned when a session cannot be located by ID.
var ErrNotFound = errors.New("session not found")

// SessionMeta is the lightweight header for a chat session — enough to
// render in a list picker without loading every message.
type SessionMeta struct {
	ID        int64
	Title     string
	Mode      int
	Model     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Store owns a SQLite database file holding all persisted sessions.
// A Store is safe to use from a single goroutine (the TUI main loop).
type Store struct {
	db *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS sessions (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	title      TEXT    NOT NULL,
	mode       INTEGER NOT NULL,
	model      TEXT    NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id   INTEGER NOT NULL,
	role         TEXT    NOT NULL,
	content      TEXT    NOT NULL,
	name         TEXT,
	tool_call_id TEXT,
	tool_calls   TEXT,
	created_at   INTEGER NOT NULL,
	FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON sessions(updated_at DESC);
`

// Open returns a Store backed by the given SQLite file path, creating
// parent directories and the schema if they are missing.
func Open(path string) (*Store, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}

	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("exec schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close releases the underlying SQLite handle.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// CreateSession inserts a new session row and returns its ID.
func (s *Store) CreateSession(title string, mode int, model string) (int64, error) {
	now := time.Now().Unix()
	res, err := s.db.Exec(
		`INSERT INTO sessions (title, mode, model, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		title, mode, model, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("insert session: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

// AppendMessage persists a single chat message and bumps the parent
// session's updated_at timestamp. Tool calls are serialized as JSON.
func (s *Store) AppendMessage(sessionID int64, msg openai.ChatCompletionMessage) error {
	var toolCallsJSON sql.NullString
	if len(msg.ToolCalls) > 0 {
		raw, err := json.Marshal(msg.ToolCalls)
		if err != nil {
			return fmt.Errorf("marshal tool calls: %w", err)
		}
		toolCallsJSON = sql.NullString{String: string(raw), Valid: true}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().Unix()
	if _, err := tx.Exec(
		`INSERT INTO messages (session_id, role, content, name, tool_call_id, tool_calls, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sessionID, msg.Role, msg.Content, nullableString(msg.Name),
		nullableString(msg.ToolCallID), toolCallsJSON, now,
	); err != nil {
		return fmt.Errorf("insert message: %w", err)
	}
	if _, err := tx.Exec(
		`UPDATE sessions SET updated_at = ? WHERE id = ?`,
		now, sessionID,
	); err != nil {
		return fmt.Errorf("bump updated_at: %w", err)
	}
	return tx.Commit()
}

// LoadSession returns the session meta plus all messages in insertion
// order. Returns ErrNotFound when no such session exists.
func (s *Store) LoadSession(sessionID int64) (SessionMeta, []openai.ChatCompletionMessage, error) {
	var meta SessionMeta
	var createdUnix, updatedUnix int64
	err := s.db.QueryRow(
		`SELECT id, title, mode, model, created_at, updated_at FROM sessions WHERE id = ?`,
		sessionID,
	).Scan(&meta.ID, &meta.Title, &meta.Mode, &meta.Model, &createdUnix, &updatedUnix)
	if errors.Is(err, sql.ErrNoRows) {
		return SessionMeta{}, nil, ErrNotFound
	}
	if err != nil {
		return SessionMeta{}, nil, fmt.Errorf("select session: %w", err)
	}
	meta.CreatedAt = time.Unix(createdUnix, 0)
	meta.UpdatedAt = time.Unix(updatedUnix, 0)

	rows, err := s.db.Query(
		`SELECT role, content, name, tool_call_id, tool_calls
		 FROM messages WHERE session_id = ? ORDER BY id ASC`,
		sessionID,
	)
	if err != nil {
		return SessionMeta{}, nil, fmt.Errorf("select messages: %w", err)
	}
	defer rows.Close()

	var msgs []openai.ChatCompletionMessage
	for rows.Next() {
		var (
			role, content string
			name, tcid    sql.NullString
			toolCallsJSON sql.NullString
		)
		if err := rows.Scan(&role, &content, &name, &tcid, &toolCallsJSON); err != nil {
			return SessionMeta{}, nil, fmt.Errorf("scan message: %w", err)
		}
		msg := openai.ChatCompletionMessage{
			Role:    role,
			Content: content,
		}
		if name.Valid {
			msg.Name = name.String
		}
		if tcid.Valid {
			msg.ToolCallID = tcid.String
		}
		if toolCallsJSON.Valid && toolCallsJSON.String != "" {
			var calls []openai.ToolCall
			if err := json.Unmarshal([]byte(toolCallsJSON.String), &calls); err != nil {
				return SessionMeta{}, nil, fmt.Errorf("unmarshal tool calls: %w", err)
			}
			msg.ToolCalls = calls
		}
		msgs = append(msgs, msg)
	}
	if err := rows.Err(); err != nil {
		return SessionMeta{}, nil, fmt.Errorf("rows err: %w", err)
	}
	return meta, msgs, nil
}

// ListSessions returns up to `limit` most recently updated sessions.
func (s *Store) ListSessions(limit int) ([]SessionMeta, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(
		`SELECT id, title, mode, model, created_at, updated_at
		 FROM sessions ORDER BY updated_at DESC, id DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var out []SessionMeta
	for rows.Next() {
		var m SessionMeta
		var c, u int64
		if err := rows.Scan(&m.ID, &m.Title, &m.Mode, &m.Model, &c, &u); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		m.CreatedAt = time.Unix(c, 0)
		m.UpdatedAt = time.Unix(u, 0)
		out = append(out, m)
	}
	return out, rows.Err()
}

// DeleteSession removes a session and, via ON DELETE CASCADE, all of
// its messages.
func (s *Store) DeleteSession(sessionID int64) error {
	res, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateSessionTitle renames a session (useful after the first user
// message arrives and we want a meaningful label).
func (s *Store) UpdateSessionTitle(sessionID int64, title string) error {
	res, err := s.db.Exec(
		`UPDATE sessions SET title = ?, updated_at = ? WHERE id = ?`,
		title, time.Now().Unix(), sessionID,
	)
	if err != nil {
		return fmt.Errorf("update title: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func nullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
