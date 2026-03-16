package engine

import (
	"context"
	"fmt"
	"time"

	"github-notifier/internal/db"
	"github-notifier/internal/github"
	"github-notifier/internal/prcomments"
)

// Source representa una fuente de notificaciones
type Source string

const (
	SourceNotification Source = "notification"
	SourcePRComment    Source = "pr_comment"
)

// UnifiedNotification representa una notificación unificada de cualquier fuente
type UnifiedNotification struct {
	ID        string
	Source    Source
	Repo      string
	Title     string
	Body      string
	Author    string
	URL       string
	Reason    string
	CreatedAt time.Time
	Resolved  bool
}

// Engine unifica notificaciones de múltiples fuentes
type Engine struct {
	notifClient *github.Client
	prClient    *prcomments.Client
	db          *db.DB
	user        string
}

// New crea un nuevo engine
func New(token, user string, database *db.DB) *Engine {
	return &Engine{
		notifClient: github.New(token),
		prClient:    prcomments.New(token, user),
		db:          database,
		user:        user,
	}
}

// FetchAll obtiene notificaciones de todas las fuentes
func (e *Engine) FetchAll(ctx context.Context) ([]*UnifiedNotification, error) {
	var all []*UnifiedNotification

	// 1. Notificaciones de GitHub (originales)
	notifs, err := e.fetchNotifications(ctx)
	if err != nil {
		// Log error pero continuar
	}
	all = append(all, notifs...)

	// 2. Comentarios de PRs
	comments, err := e.fetchPRComments(ctx)
	if err != nil {
		// Log error pero continuar
	}
	all = append(all, comments...)

	return all, nil
}

func (e *Engine) fetchNotifications(ctx context.Context) ([]*UnifiedNotification, error) {
	notifications, err := e.notifClient.FetchNotifications(ctx)
	if err != nil {
		return nil, err
	}

	var result []*UnifiedNotification
	for _, n := range notifications {
		result = append(result, &UnifiedNotification{
			ID:        n.ID,
			Source:    SourceNotification,
			Repo:      n.Repo,
			Title:     n.Title,
			Body:      "",
			Author:    "",
			URL:       n.URL,
			Reason:    n.Reason,
			CreatedAt: n.UpdatedAt,
			Resolved:  false,
		})
	}

	return result, nil
}

func (e *Engine) fetchPRComments(ctx context.Context) ([]*UnifiedNotification, error) {
	comments, err := e.prClient.FetchComments(ctx)
	if err != nil {
		return nil, err
	}

	var result []*UnifiedNotification
	for _, c := range comments {
		// Crear ID único para comentarios de PR
		id := fmt.Sprintf("pr-comment-%d-%d", c.PRNumber, c.ID)

		result = append(result, &UnifiedNotification{
			ID:        id,
			Source:    SourcePRComment,
			Repo:      c.Repo,
			Title:     fmt.Sprintf("#%d: %s", c.PRNumber, c.PRTitle),
			Body:      c.Body,
			Author:    c.Author,
			URL:       c.URL,
			Reason:    c.Type,
			CreatedAt: c.CreatedAt,
			Resolved:  false,
		})
	}

	return result, nil
}

// Persist guarda las notificaciones en la base de datos
func (e *Engine) Persist(notifications []*UnifiedNotification) error {
	for _, n := range notifications {
		// Verificar si ya existe
		existing, err := e.db.GetByID(n.ID)
		if err != nil {
			continue
		}

		if existing == nil {
			// Nueva notificación
			dbNotif := &db.Notification{
				ID:        n.ID,
				Repo:      n.Repo,
				Title:     n.Title,
				Type:      string(n.Source),
				URL:       n.URL,
				Reason:    n.Reason,
				Author:    n.Author,
				Unread:    true,
				Resolved:  false,
				CreatedAt: time.Now(),
				UpdatedAt: n.CreatedAt,
			}
			if err := e.db.Upsert(dbNotif); err != nil {
				continue
			}
		}
	}
	return nil
}

// GetUnresolved obtiene todas las notificaciones no resueltas
func (e *Engine) GetUnresolved() ([]*db.Notification, error) {
	return e.db.ListUnresolved()
}

// MarkResolved marca una notificación como resuelta
func (e *Engine) MarkResolved(id string) error {
	return e.db.MarkResolved(id)
}
