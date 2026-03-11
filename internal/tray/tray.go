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
	ghclient "github-notifier/internal/github"
)

const maxItems = 15

// App manages the system tray lifecycle.
type App struct {
	cfg    *config.Config
	db     *db.DB
	github *ghclient.Client

	mu        sync.Mutex
	items     []*db.Notification
	menuItems []*systray.MenuItem

	mCount   *systray.MenuItem
	mRefresh *systray.MenuItem
}

func New(cfg *config.Config, database *db.DB) *App {
	return &App{
		cfg:    cfg,
		db:     database,
		github: ghclient.New(cfg.GitHubToken),
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
		item := systray.AddMenuItem("", "")
		item.Hide()
		a.menuItems = append(a.menuItems, item)
	}

	systray.AddSeparator()
	a.mRefresh = systray.AddMenuItem("Actualizar", "Buscar nuevas notificaciones")
	mMarkAll := systray.AddMenuItem("Marcar todo como leído", "Resolver todas las notificaciones")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Salir", "Cerrar GitHub Notifier")

	// Start background polling.
	go a.poll()

	// Handle notification item clicks in individual goroutines.
	for i, item := range a.menuItems {
		go a.handleClick(i, item)
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
	// Use a ticker driven by the channel approach to avoid importing time in a goroutine race.
	ticker := newTicker(a.cfg.PollInterval)
	defer ticker.stop()
	for range ticker.ch {
		a.refresh()
	}
}

// refresh fetches from GitHub, persists, sends desktop notifications, updates menu.
func (a *App) refresh() {
	ctx := context.Background()
	notifications, err := a.github.FetchNotifications(ctx)
	if err != nil {
		log.Printf("Error al obtener notificaciones: %v", err)
		return
	}

	for _, n := range notifications {
		existing, _ := a.db.GetByID(n.ID)
		if existing == nil {
			// Brand-new notification — send desktop alert.
			sendDesktopNotification(n)
		}
		if err := a.db.Upsert(n); err != nil {
			log.Printf("Error al guardar notificación %s: %v", n.ID, err)
		}
	}

	a.updateMenu()
}

// updateMenu rebuilds the tray menu from the DB.
func (a *App) updateMenu() {
	unresolved, err := a.db.ListUnresolved()
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
			label := fmt.Sprintf("%s  %s", reasonEmoji(n.Reason), truncate(n.Title, 42))
			slot.SetTitle(label)
			slot.SetTooltip(fmt.Sprintf("%s\n%s\nClick para abrir el comentario y resolver", n.Repo, humanReason(n.Reason)))
			slot.Show()
		} else {
			slot.Hide()
		}
	}
}

// handleClick opens the notification URL and marks it resolved when clicked.
func (a *App) handleClick(idx int, item *systray.MenuItem) {
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

		if err := a.db.MarkResolved(n.ID); err != nil {
			log.Printf("Error al marcar como resuelto: %v", err)
		}
		if err := a.github.MarkRead(context.Background(), n.ID); err != nil {
			log.Printf("Error al marcar como leído en GitHub: %v", err)
		}

		go a.updateMenu()
	}
}

// markAllRead resolves every pending notification.
func (a *App) markAllRead() {
	unresolved, err := a.db.ListUnresolved()
	if err != nil {
		return
	}
	ctx := context.Background()
	for _, n := range unresolved {
		_ = a.db.MarkResolved(n.ID)
		_ = a.github.MarkRead(ctx, n.ID)
	}
	a.updateMenu()
}

// sendDesktopNotification calls notify-send to show a system notification.
func sendDesktopNotification(n *db.Notification) {
	summary := fmt.Sprintf("GitHub: %s", n.Reason)
	body := fmt.Sprintf("[%s]\n%s", n.Repo, n.Title)
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
	default:
		return reason
	}
}
