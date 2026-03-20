package db

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

type Session struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID        int64     `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}

	return db, nil
}

func (d *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON sessions(updated_at);
	`

	_, err := d.conn.Exec(schema)
	return err
}

func (d *DB) Close() error {
	return d.conn.Close()
}

// Session operations
func (d *DB) CreateSession(id, title string) (*Session, error) {
	now := time.Now()
	session := &Session{
		ID:        id,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := d.conn.Exec(
		"INSERT INTO sessions (id, title, created_at, updated_at) VALUES (?, ?, ?, ?)",
		session.ID, session.Title, session.CreatedAt, session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (d *DB) GetSession(id string) (*Session, error) {
	session := &Session{}
	err := d.conn.QueryRow(
		"SELECT id, title, created_at, updated_at FROM sessions WHERE id = ?",
		id,
	).Scan(&session.ID, &session.Title, &session.CreatedAt, &session.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (d *DB) ListSessions(limit int) ([]*Session, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := d.conn.Query(
		"SELECT id, title, created_at, updated_at FROM sessions ORDER BY updated_at DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		session := &Session{}
		if err := rows.Scan(&session.ID, &session.Title, &session.CreatedAt, &session.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

func (d *DB) UpdateSessionTitle(id, title string) error {
	_, err := d.conn.Exec(
		"UPDATE sessions SET title = ?, updated_at = ? WHERE id = ?",
		title, time.Now(), id,
	)
	return err
}

func (d *DB) UpdateSessionTimestamp(id string) error {
	_, err := d.conn.Exec(
		"UPDATE sessions SET updated_at = ? WHERE id = ?",
		time.Now(), id,
	)
	return err
}

func (d *DB) DeleteSession(id string) error {
	_, err := d.conn.Exec("DELETE FROM sessions WHERE id = ?", id)
	return err
}

// Message operations
func (d *DB) CreateMessage(sessionID, role, content string) (*Message, error) {
	result, err := d.conn.Exec(
		"INSERT INTO messages (session_id, role, content, created_at) VALUES (?, ?, ?, ?)",
		sessionID, role, content, time.Now(),
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &Message{
		ID:        id,
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}, nil
}

func (d *DB) GetMessagesBySession(sessionID string, limit int) ([]*Message, error) {
	if limit <= 0 {
		limit = 1000
	}

	rows, err := d.conn.Query(
		"SELECT id, session_id, role, content, created_at FROM messages WHERE session_id = ? ORDER BY created_at ASC LIMIT ?",
		sessionID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

func (d *DB) DeleteMessagesBySession(sessionID string) error {
	_, err := d.conn.Exec("DELETE FROM messages WHERE session_id = ?", sessionID)
	return err
}
