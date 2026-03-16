package tray

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"sync"

	"github.com/getlantern/systray"

	"github-notifier/icon"
	"github-notifier/internal/config"
	"github-notifier/internal/db"
	"github-notifier/internal/engine"
)

const maxItems = 15

// App manages the system tray lifecycle.
type App struct {
	cfg    *config.Config
	db     *db.DB
	engine *engine.Engine

	mu        sync.Mutex
	items     []*db.Notification
	menuItems []*menuItem

	mCount   *systray.MenuItem
	mRefresh *systray.MenuItem
}

type menuItem struct {
	parent  *systray.MenuItem
	open    *systray.MenuItem
	resolve *systray.MenuItem
}

func New(cfg *config.Config, database *db.DB) *App {
	return &App{
		cfg:    cfg,
		db:     database,
		engine: engine.New(cfg.GitHubToken, cfg.GitHubUser, database),
	}
}

func (a *App) OnReady() {
	systray.SetIcon(icon.Normal())
	systray.SetTooltip("GitHub Notifier")

	a.mCount = systray.AddMenuItem("Cargando notificaciones...", "")
	a.mCount.Disable()
	systray.AddSeparator()

	// Pre-create hidden slots for notification entries.
	for i := 0; i < maxItems; i++ {
		parent := systray.AddMenuItem("", "")
		parent.Hide()

		// Submenu items
		open := parent.AddSubMenuItem("Abrir", "Abrir el comentario en el navegador")
		resolve := parent.AddSubMenuItem("Marcar como resuelto", "Marcar esta notificación como resuelta")

		a.menuItems = append(a.menuItems, &menuItem{
			parent:  parent,
			open:    open,
			resolve: resolve,
		})
	}

	systray.AddSeparator()
	a.mRefresh = systray.AddMenuItem("Actualizar", "Buscar nuevas notificaciones")
	mMarkAll := systray.AddMenuItem("Marcar todo como resuelto", "Resolver todas las notificaciones")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Salir", "Cerrar GitHub Notifier")

	// Start background polling.
	go a.poll()

	// Handle notification item clicks in individual goroutines.
	for i, item := range a.menuItems {
		go a.handleOpenClick(i, item.open)
		go a.handleResolveClick(i, item.resolve)
	}

	// Handle control menu events.
	go func() {
		for {
			select {
			case <-a.mRefresh.ClickedCh:
				go a.refresh()
			case <-mMarkAll.ClickedCh:
				go a.markAllRead()
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func (a *App) OnExit() {
	log.Println("GitHub Notifier cerrado.")
}

// poll runs the refresh loop on the configured interval.
func (a *App) poll() {
	a.refresh()
	ticker := newTicker(a.cfg.PollInterval)
	defer ticker.stop()
	for range ticker.ch {
		a.refresh()
	}
}

// refresh fetches from all sources, persists, sends desktop notifications, updates menu.
func (a *App) refresh() {
	ctx := context.Background()

	// Fetch all notifications from all sources
	notifications, err := a.engine.FetchAll(ctx)
	if err != nil {
		log.Printf("Error al obtener notificaciones: %v", err)
		return
	}

	// Persist to database
	if err := a.engine.Persist(notifications); err != nil {
		log.Printf("Error al persistir notificaciones: %v", err)
	}

	// Send desktop notifications for new ones
	for _, n := range notifications {
		existing, _ := a.db.GetByID(n.ID)
		if existing == nil {
			sendDesktopNotification(n)
		}
	}

	a.updateMenu()
}

// updateMenu rebuilds the tray menu from the DB.
func (a *App) updateMenu() {
	unresolved, err := a.engine.GetUnresolved()
	if err != nil {
		log.Printf("Error al listar notificaciones: %v", err)
		return
	}

	a.mu.Lock()
	a.items = unresolved
	a.mu.Unlock()

	count := len(unresolved)
	if count == 0 {
		systray.SetIcon(icon.Normal())
		a.mCount.SetTitle("Sin comentarios pendientes en PRs")
	} else {
		systray.SetIcon(icon.Alert())
		a.mCount.SetTitle(fmt.Sprintf("%d comentario(s) de PR sin resolver", count))
	}

	for i, slot := range a.menuItems {
		if i < len(unresolved) {
			n := unresolved[i]
			// Label: Emoji + Autor + Título truncado
			label := fmt.Sprintf("%s %s: %s", reasonEmoji(n.Reason), truncate(n.Author, 12), truncate(n.Title, 35))
			slot.parent.SetTitle(label)

			// Tooltip con toda la metadata
			tooltip := fmt.Sprintf("📁 %s\n👤 %s\n📝 %s\n\n💬 %s",
				n.Repo,
				n.Author,
				humanReason(n.Reason),
				truncate(n.Title, 60))
			slot.parent.SetTooltip(tooltip)
			slot.parent.Show()
		} else {
			slot.parent.Hide()
		}
	}
}

// handleOpenClick opens the notification URL when clicked.
func (a *App) handleOpenClick(idx int, item *systray.MenuItem) {
	for range item.ClickedCh {
		a.mu.Lock()
		if idx >= len(a.items) {
			a.mu.Unlock()
			continue
		}
		n := a.items[idx]
		a.mu.Unlock()

		if n.URL != "" {
			if err := exec.Command("xdg-open", n.URL).Start(); err != nil {
				log.Printf("Error al abrir URL: %v", err)
			}
		}
	}
}

// handleResolveClick marks the notification as resolved.
func (a *App) handleResolveClick(idx int, item *systray.MenuItem) {
	for range item.ClickedCh {
		a.mu.Lock()
		if idx >= len(a.items) {
			a.mu.Unlock()
			continue
		}
		n := a.items[idx]
		a.mu.Unlock()

		if err := a.engine.MarkResolved(n.ID); err != nil {
			log.Printf("Error al marcar como resuelto: %v", err)
		}

		go a.updateMenu()
	}
}

// markAllRead resolves every pending notification.
func (a *App) markAllRead() {
	unresolved, err := a.engine.GetUnresolved()
	if err != nil {
		return
	}
	for _, n := range unresolved {
		_ = a.engine.MarkResolved(n.ID)
	}
	a.updateMenu()
}

// sendDesktopNotification calls notify-send to show a system notification.
func sendDesktopNotification(n *engine.UnifiedNotification) {
	summary := fmt.Sprintf("GitHub: %s", n.Reason)
	body := fmt.Sprintf("[%s]\n%s", n.Repo, n.Title)
	if n.Author != "" {
		body = fmt.Sprintf("[%s] %s: %s", n.Repo, n.Author, n.Title)
	}
	cmd := exec.Command("notify-send",
		"--icon=dialog-information",
		"--urgency=normal",
		"--expire-time=8000",
		summary, body,
	)
	if err := cmd.Run(); err != nil {
		log.Printf("notify-send error: %v", err)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func reasonEmoji(reason string) string {
	switch reason {
	case "comment":
		return "💬"
	case "mention", "team_mention":
		return "📢"
	case "review_requested":
		return "👀"
	case "author":
		return "✏️"
	case "subscribed":
		return "🔔"
	case "manual":
		return "📌"
	case "review_comment":
		return "📝"
	case "issue_comment":
		return "💭"
	default:
		return "•"
	}
}

func humanReason(reason string) string {
	switch reason {
	case "comment":
		return "Nuevo comentario en tu PR"
	case "mention":
		return "Te mencionaron en un comentario"
	case "team_mention":
		return "Mencionaron a tu equipo"
	case "review_requested":
		return "Te pidieron review"
	case "author":
		return "Actividad en un PR tuyo"
	case "subscribed":
		return "Actividad en PR suscripto"
	case "manual":
		return "Notificación manual"
	case "review_comment":
		return "Comentario en código"
	case "issue_comment":
		return "Comentario en PR"
	default:
		return reason
	}
}
