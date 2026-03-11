package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Notification struct {
	ID        string
	Repo      string
	Title     string
	Type      string
	URL       string
	Reason    string
	Unread    bool
	Resolved  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DB struct {
	db *sql.DB
}

func New(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS notifications (
			id         TEXT PRIMARY KEY,
			repo       TEXT NOT NULL,
			title      TEXT NOT NULL,
			type       TEXT NOT NULL,
			url        TEXT NOT NULL,
			reason     TEXT NOT NULL,
			unread     INTEGER NOT NULL DEFAULT 1,
			resolved   INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`)
	return err
}

func (d *DB) Upsert(n *Notification) error {
	_, err := d.db.Exec(`
		INSERT INTO notifications (id, repo, title, type, url, reason, unread, resolved, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title      = excluded.title,
			unread     = excluded.unread,
			updated_at = excluded.updated_at
	`,
		n.ID, n.Repo, n.Title, n.Type, n.URL, n.Reason,
		boolInt(n.Unread), boolInt(n.Resolved),
		n.CreatedAt, n.UpdatedAt,
	)
	return err
}

func (d *DB) GetByID(id string) (*Notification, error) {
	row := d.db.QueryRow(`
		SELECT id, repo, title, type, url, reason, unread, resolved, created_at, updated_at
		FROM notifications WHERE id = ?
	`, id)

	n := &Notification{}
	var unread, resolved int
	err := row.Scan(&n.ID, &n.Repo, &n.Title, &n.Type, &n.URL, &n.Reason,
		&unread, &resolved, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	n.Unread = unread == 1
	n.Resolved = resolved == 1
	return n, nil
}

func (d *DB) ListUnresolved() ([]*Notification, error) {
	rows, err := d.db.Query(`
		SELECT id, repo, title, type, url, reason, unread, resolved, created_at, updated_at
		FROM notifications
		WHERE resolved = 0 AND unread = 1
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Notification
	for rows.Next() {
		n := &Notification{}
		var unread, resolved int
		if err := rows.Scan(&n.ID, &n.Repo, &n.Title, &n.Type, &n.URL, &n.Reason,
			&unread, &resolved, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		n.Unread = unread == 1
		n.Resolved = resolved == 1
		list = append(list, n)
	}
	return list, rows.Err()
}

func (d *DB) MarkResolved(id string) error {
	_, err := d.db.Exec(`UPDATE notifications SET resolved = 1, unread = 0 WHERE id = ?`, id)
	return err
}

func (d *DB) Close() error {
	return d.db.Close()
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
