package db

import (
	"testing"
	"time"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := New(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func sampleNotification(id string) *Notification {
	return &Notification{
		ID:        id,
		Repo:      "alice/myrepo",
		Title:     "Fix the bug in the login flow",
		Type:      "PullRequest",
		URL:       "https://github.com/alice/myrepo/pull/1#issuecomment-123",
		Reason:    "comment",
		Author:    "testuser",
		Unread:    true,
		Resolved:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestUpsertAndGetByID(t *testing.T) {
	db := newTestDB(t)
	n := sampleNotification("abc123")

	if err := db.Upsert(n); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := db.GetByID("abc123")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil {
		t.Fatal("expected notification, got nil")
	}
	if got.Title != n.Title {
		t.Errorf("Title = %q, want %q", got.Title, n.Title)
	}
	if got.Repo != n.Repo {
		t.Errorf("Repo = %q, want %q", got.Repo, n.Repo)
	}
	if got.Author != n.Author {
		t.Errorf("Author = %q, want %q", got.Author, n.Author)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	db := newTestDB(t)
	got, err := db.GetByID("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestUpsert_UpdatesTitle(t *testing.T) {
	db := newTestDB(t)
	n := sampleNotification("upd1")
	db.Upsert(n)

	n.Title = "Updated title"
	db.Upsert(n)

	got, _ := db.GetByID("upd1")
	if got.Title != "Updated title" {
		t.Errorf("expected updated title, got %q", got.Title)
	}
}

func TestListUnresolved(t *testing.T) {
	db := newTestDB(t)

	db.Upsert(sampleNotification("n1"))
	db.Upsert(sampleNotification("n2"))

	resolved := sampleNotification("n3")
	resolved.Resolved = true
	db.Upsert(resolved)

	list, err := db.ListUnresolved()
	if err != nil {
		t.Fatalf("ListUnresolved: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 unresolved, got %d", len(list))
	}
}

func TestMarkResolved(t *testing.T) {
	db := newTestDB(t)
	db.Upsert(sampleNotification("resolve-me"))

	if err := db.MarkResolved("resolve-me"); err != nil {
		t.Fatalf("MarkResolved: %v", err)
	}

	got, _ := db.GetByID("resolve-me")
	if !got.Resolved {
		t.Error("expected Resolved = true")
	}
	if got.Unread {
		t.Error("expected Unread = false after resolving")
	}
}

func TestListUnresolved_EmptyAfterMarkAll(t *testing.T) {
	db := newTestDB(t)
	db.Upsert(sampleNotification("a"))
	db.Upsert(sampleNotification("b"))

	db.MarkResolved("a")
	db.MarkResolved("b")

	list, _ := db.ListUnresolved()
	if len(list) != 0 {
		t.Errorf("expected 0 after marking all resolved, got %d", len(list))
	}
}
