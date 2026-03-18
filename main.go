package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/getlantern/systray"

	"github-notifier/internal/config"
	"github-notifier/internal/db"
	"github-notifier/internal/tray"
	"github-notifier/internal/updater"
)

// Set via -ldflags at build time.
var Version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version":
			fmt.Println(Version)
			return
		case "update":
			runSelfUpdate()
			return
		}
	}

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

	// Background update check every 6h; notify via desktop notification.
	updater.StartBackgroundCheck(Version, func(result *updater.CheckResult) {
		tray.SendUpdateNotification(result.Current, result.Latest)
	})

	// Graceful shutdown on SIGINT / SIGTERM.
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		systray.Quit()
	}()

	systray.Run(app.OnReady, app.OnExit)
}

func runSelfUpdate() {
	fmt.Printf("Current version: %s\n", Version)
	fmt.Println("Checking for updates...")

	result, err := updater.Check(Version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !result.Available {
		fmt.Println("Already at latest version.")
		return
	}

	fmt.Printf("New version available: %s\n", result.Latest)
	fmt.Println("Downloading and applying update...")

	if err := updater.Apply(result.Latest); err != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Update applied. Service restarting.")
}
