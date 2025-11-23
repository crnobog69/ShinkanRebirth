package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GotifyServer     string
	GotifyToken      string
	DiscordToken     string
	DiscordChannelID string
	WebPort          string
	CheckInterval    string
	MangaDataFile    string
	AnimeDataFile    string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		GotifyServer:     getEnv("GOTIFY_SERVER", ""),
		GotifyToken:      getEnv("GOTIFY_TOKEN", ""),
		DiscordToken:     getEnv("DISCORD_TOKEN", ""),
		DiscordChannelID: getEnv("DISCORD_CHANNEL_ID", ""),
		WebPort:          getEnv("WEB_PORT", "11111"),
		CheckInterval:    getEnv("CHECK_INTERVAL", "0 * * * *"),
		MangaDataFile:    getEnv("MANGA_DATA_FILE", "./data/mangas.json"),
		AnimeDataFile:    getEnv("ANIME_DATA_FILE", "./data/anime.json"),
	}

	// Validate required configuration (at least one notification method)
	if cfg.GotifyServer == "" && cfg.DiscordToken == "" {
		log.Fatal("❌ ERROR: Either GOTIFY_SERVER or DISCORD_TOKEN must be set in .env file")
	}

	if cfg.GotifyServer != "" && cfg.GotifyToken == "" {
		log.Fatal("❌ ERROR: GOTIFY_TOKEN is required when GOTIFY_SERVER is set")
	}

	if cfg.DiscordToken != "" && cfg.DiscordChannelID == "" {
		log.Fatal("❌ ERROR: DISCORD_CHANNEL_ID is required when DISCORD_TOKEN is set")
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
