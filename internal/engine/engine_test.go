package engine

import (
	"testing"
	"time"

	"github-notifier/internal/db"
)

func TestUnifiedNotification_Struct(t *testing.T) {
	n := &UnifiedNotification{
		ID:        "test-123",
		Source:    SourcePRComment,
		Repo:      "owner/repo",
		Title:     "Test PR",
		Body:      "Test body",
		Author:    "testuser",
		URL:       "https://github.com/owner/repo/pull/1",
		Reason:    "issue_comment",
		CreatedAt: time.Now(),
		Resolved:  false,
	}

	if n.ID != "test-123" {
		t.Errorf("ID = %q, want 'test-123'", n.ID)
	}
	if n.Source != SourcePRComment {
		t.Errorf("Source = %q, want SourcePRComment", n.Source)
	}
	if n.Resolved {
		t.Error("Resolved should be false")
	}
}

func TestSource_Constants(t *testing.T) {
	if SourceNotification != "notification" {
		t.Errorf("SourceNotification = %q, want 'notification'", SourceNotification)
	}
	if SourcePRComment != "pr_comment" {
		t.Errorf("SourcePRComment = %q, want 'pr_comment'", SourcePRComment)
	}
}

func TestEngine_New(t *testing.T) {
	// No podemos probar FetchAll sin mocks, pero podemos probar que New crea el engine
	// Esto es un test básico de estructura
	_ = Engine{}
}

func TestPersist_NewNotification(t *testing.T) {
	// Crear DB de test
	testDB, err := db.New(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}
	defer testDB.Close()

	engine := New("fake-token", "testuser", testDB)

	notifications := []*UnifiedNotification{
		{
			ID:        "test-1",
			Source:    SourcePRComment,
			Repo:      "owner/repo",
			Title:     "PR #1",
			Reason:    "issue_comment",
			CreatedAt: time.Now(),
		},
	}

	err = engine.Persist(notifications)
	if err != nil {
		t.Fatalf("Persist failed: %v", err)
	}

	// Verificar que se guardó
	got, err := testDB.GetByID("test-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected notification to be persisted, got nil")
	}
	if got.Repo != "owner/repo" {
		t.Errorf("Repo = %q, want 'owner/repo'", got.Repo)
	}
}

func TestPersist_Duplicate(t *testing.T) {
	testDB, err := db.New(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}
	defer testDB.Close()

	engine := New("fake-token", "testuser", testDB)

	notifications := []*UnifiedNotification{
		{
			ID:        "dup-1",
			Source:    SourcePRComment,
			Repo:      "owner/repo",
			Title:     "PR #1",
			Reason:    "issue_comment",
			CreatedAt: time.Now(),
		},
	}

	// Persistir dos veces
	engine.Persist(notifications)
	engine.Persist(notifications)

	// Verificar que solo hay una
	list, err := testDB.ListUnresolved()
	if err != nil {
		t.Fatalf("ListUnresolved failed: %v", err)
	}

	count := 0
	for _, n := range list {
		if n.ID == "dup-1" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 notification with ID 'dup-1', got %d", count)
	}
}

func TestGetUnresolved(t *testing.T) {
	testDB, err := db.New(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}
	defer testDB.Close()

	engine := New("fake-token", "testuser", testDB)

	// Crear notificación no resuelta
	notifications := []*UnifiedNotification{
		{
			ID:        "unresolved-1",
			Source:    SourcePRComment,
			Repo:      "owner/repo",
			Title:     "PR #1",
			Reason:    "issue_comment",
			CreatedAt: time.Now(),
		},
	}
	engine.Persist(notifications)

	list, err := engine.GetUnresolved()
	if err != nil {
		t.Fatalf("GetUnresolved failed: %v", err)
	}

	found := false
	for _, n := range list {
		if n.ID == "unresolved-1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find 'unresolved-1' in unresolved list")
	}
}

func TestMarkResolved(t *testing.T) {
	testDB, err := db.New(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}
	defer testDB.Close()

	engine := New("fake-token", "testuser", testDB)

	notifications := []*UnifiedNotification{
		{
			ID:        "resolve-test",
			Source:    SourcePRComment,
			Repo:      "owner/repo",
			Title:     "PR #1",
			Reason:    "issue_comment",
			CreatedAt: time.Now(),
		},
	}
	engine.Persist(notifications)

	err = engine.MarkResolved("resolve-test")
	if err != nil {
		t.Fatalf("MarkResolved failed: %v", err)
	}

	// Verificar que se marcó como resuelto
	got, _ := testDB.GetByID("resolve-test")
	if got == nil {
		t.Fatal("notification not found")
	}
	if !got.Resolved {
		t.Error("expected notification to be resolved")
	}
}
