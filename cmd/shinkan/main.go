package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shinkan-rebirth/internal/checker"
	"shinkan-rebirth/internal/config"
	"shinkan-rebirth/internal/notifier"
	"shinkan-rebirth/internal/quotes"
	"shinkan-rebirth/internal/storage"
	"shinkan-rebirth/internal/web"

	"github.com/robfig/cron/v3"
)

const banner = `
===================================================
  _________.__    .__        __                   
 /   _____/|  |__ |__| ____ |  | _______    ____  
 \_____  \ |  |  \|  |/    \|  |/ /\__  \  /    \ 
 /        \|   Y  \  |   |  \    <  / __ \|   |  \
/_______  /|___|  /__|___|  /__|_ \(____  /___|  /
        \/      \/        \/     \/     \/     \/ 
        Rebirth - Go Edition with Gotify
===================================================
`

func main() {
	fmt.Println(banner)

	// Load configuration
	cfg := config.Load()

	// Initialize components
	store := storage.New(cfg.MangaDataFile, cfg.AnimeDataFile)
	notify := notifier.New(cfg.GotifyServer, cfg.GotifyToken, cfg.DiscordToken, cfg.DiscordChannelID)
	check := checker.New(store, notify)
	quoteManager, err := quotes.New("./data/quotes.json")
	if err != nil {
		log.Printf("⚠️ Failed to load quotes: %v\n", err)
	}
	
	// Register Discord slash commands
	if err := notify.RegisterCommands(
		func() { check.CheckAll() },
		func() string { 
			if quoteManager != nil {
				return quoteManager.GetRandom()
			}
			return "💭 No quotes available."
		},
	); err != nil {
		log.Printf("⚠️ Failed to register commands: %v\n", err)
	}
	
	// Ensure notifier cleanup on exit
	defer notify.Close()

	// Track start time
	startTime := time.Now()

	// Start web server in goroutine
	server := web.New(store, check, startTime)
	go func() {
		if err := server.Start(cfg.WebPort); err != nil {
			log.Fatalf("❌ Failed to start web server: %v", err)
		}
	}()

	log.Println("✅ Shinkan Rebirth initialized")

	// Initial check
	log.Println("🔍 Starting initial feed check...")
	check.CheckAll()

	// Setup cron scheduler
	c := cron.New()
	_, cronErr := c.AddFunc(cfg.CheckInterval, func() {
		log.Println("⏰ Scheduled check triggered")
		check.CheckAll()
	})

	if cronErr != nil {
		log.Fatalf("❌ Failed to setup cron schedule: %v", cronErr)
	}

	c.Start()
	log.Printf("⏰ Scheduled checks: %s\n", cfg.CheckInterval)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("\n👋 Shutting down gracefully...")
	c.Stop()
	notify.Close()
	log.Println("✅ Goodbye!")
}
