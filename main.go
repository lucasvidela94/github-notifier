package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/getlantern/systray"

	"github-notifier/internal/config"
	"github-notifier/internal/db"
	"github-notifier/internal/tray"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuración inválida: %v", err)
	}

	database, err := db.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("No se pudo abrir la base de datos: %v", err)
	}
	defer database.Close()

	app := tray.New(cfg, database)

	// Graceful shutdown on SIGINT / SIGTERM.
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		systray.Quit()
	}()

	systray.Run(app.OnReady, app.OnExit)
}
