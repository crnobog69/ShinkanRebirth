package web

import (
	"log"
	"strings"
	"time"

	"shinkan-rebirth/internal/checker"
	"shinkan-rebirth/internal/models"
	"shinkan-rebirth/internal/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Server struct {
	app       *fiber.App
	storage   *storage.Storage
	checker   *checker.Checker
	startTime time.Time
}

func New(storage *storage.Storage, checker *checker.Checker, startTime time.Time) *Server {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Middleware
	app.Use(cors.New())
	app.Use(logger.New())

	server := &Server{
		app:       app,
		storage:   storage,
		checker:   checker,
		startTime: startTime,
	}

	server.setupRoutes()

	return server
}

func (s *Server) setupRoutes() {
	// Static files
	s.app.Static("/", "./public")

	// API routes
	api := s.app.Group("/api")

	api.Get("/feeds", s.getFeeds)
	api.Get("/categories", s.getCategories)
	api.Get("/stats", s.getStats)
	api.Get("/health", s.getHealth)
	api.Post("/feeds", s.addFeed)
	api.Post("/import", s.importFeeds)
	api.Delete("/feeds/:id", s.deleteFeed)
	api.Put("/feeds/:id", s.updateFeed)
	api.Post("/feeds/:id/test", s.testFeed)
	api.Post("/feeds/:id/check", s.checkFeed)
	api.Get("/export", s.exportFeeds)
}

func (s *Server) getFeeds(c *fiber.Ctx) error {
	category := c.Query("category")
	search := c.Query("search")

	feeds, err := s.storage.GetFeeds()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if search != "" {
		feeds, err = s.storage.SearchFeeds(search)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	} else if category != "" && category != "all" {
		filtered := make([]models.Feed, 0)
		for _, feed := range feeds {
			feedCategory := feed.Category
			if feedCategory == "" {
				feedCategory = "Uncategorized"
			}
			if feedCategory == category {
				filtered = append(filtered, feed)
			}
		}
		feeds = filtered
	}

	return c.JSON(feeds)
}

func (s *Server) getCategories(c *fiber.Ctx) error {
	categories, err := s.storage.GetCategories()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(categories)
}

func (s *Server) getStats(c *fiber.Ctx) error {
	feeds, err := s.storage.GetFeeds()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	checkerStats := s.checker.GetStats()
	categories, _ := s.storage.GetCategories()

	feedsWithErrors := 0
	feedsNeverChecked := 0
	for _, feed := range feeds {
		if feed.FailCount > 0 {
			feedsWithErrors++
		}
		if feed.LastChecked == nil {
			feedsNeverChecked++
		}
	}

	stats := models.Stats{
		TotalChecks:       checkerStats.TotalChecks,
		SuccessfulChecks:  checkerStats.SuccessfulChecks,
		FailedChecks:      checkerStats.FailedChecks,
		NotificationsSent: checkerStats.NotificationsSent,
		LastCheckTime:     checkerStats.LastCheckTime,
		TotalFeeds:        len(feeds),
		FeedsWithErrors:   feedsWithErrors,
		FeedsNeverChecked: feedsNeverChecked,
		Categories:        len(categories),
		Uptime:            time.Since(s.startTime).Milliseconds(),
	}

	return c.JSON(stats)
}

func (s *Server) getHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(s.startTime).Seconds(),
	})
}

func (s *Server) addFeed(c *fiber.Ctx) error {
	var req struct {
		Name       string  `json:"name"`
		RSSUrl     string  `json:"rssUrl"`
		Type       string  `json:"type"`
		AnilistUrl *string `json:"anilistUrl"`
		Category   string  `json:"category"`
		SearchText *string `json:"searchText"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" || req.RSSUrl == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name and RSS URL required"})
	}

	// Default to manga if type not specified
	if req.Type == "" {
		req.Type = string(models.FeedTypeManga)
	}

	// Auto-append /rss if not present (for manga feeds)
	if req.Type == string(models.FeedTypeManga) && !strings.HasSuffix(req.RSSUrl, "/rss") {
		req.RSSUrl = strings.TrimSuffix(req.RSSUrl, "/") + "/rss"
	}

	feed := models.Feed{
		Name:       req.Name,
		RSSUrl:     req.RSSUrl,
		Type:       models.FeedType(req.Type),
		AnilistUrl: req.AnilistUrl,
		Category:   req.Category,
		SearchText: req.SearchText,
	}

	newFeed, err := s.storage.AddFeed(feed)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(newFeed)
}

func (s *Server) importFeeds(c *fiber.Ctx) error {
	var req struct {
		Feeds []models.Feed `json:"feeds"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Feeds == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid import data"})
	}

	imported, skipped, err := s.storage.ImportFeeds(req.Feeds)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"imported": imported,
		"skipped":  skipped,
	})
}

func (s *Server) deleteFeed(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := s.storage.DeleteFeed(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func (s *Server) updateFeed(c *fiber.Ctx) error {
	id := c.Params("id")

	var req struct {
		Name       string  `json:"name"`
		RSSUrl     string  `json:"rssUrl"`
		Type       string  `json:"type"`
		AnilistUrl *string `json:"anilistUrl"`
		Category   string  `json:"category"`
		SearchText *string `json:"searchText"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" || req.RSSUrl == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name and RSS URL required"})
	}

	// Auto-append /rss if not present (for manga feeds)
	if req.Type == string(models.FeedTypeManga) && !strings.HasSuffix(req.RSSUrl, "/rss") {
		req.RSSUrl = strings.TrimSuffix(req.RSSUrl, "/") + "/rss"
	}

	updates := map[string]interface{}{
		"name":       req.Name,
		"rssUrl":     req.RSSUrl,
		"type":       req.Type,
		"category":   req.Category,
		"searchText": req.SearchText,
	}

	if req.AnilistUrl != nil {
		updates["anilistUrl"] = *req.AnilistUrl
	}

	feed, err := s.storage.UpdateFeed(id, updates)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Feed not found"})
	}

	return c.JSON(feed)
}

func (s *Server) testFeed(c *fiber.Ctx) error {
	id := c.Params("id")

	result, err := s.checker.TestFeed(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

func (s *Server) checkFeed(c *fiber.Ctx) error {
	id := c.Params("id")

	feeds, err := s.storage.GetFeeds()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var feed *models.Feed
	for _, f := range feeds {
		if f.ID == id {
			feed = &f
			break
		}
	}

	if feed == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Feed not found"})
	}

	if err := s.checker.CheckFeed(*feed, 3); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Get updated feed
	feeds, _ = s.storage.GetFeeds()
	for _, f := range feeds {
		if f.ID == id {
			feed = &f
			break
		}
	}

	return c.JSON(fiber.Map{
		"success":     true,
		"lastChapter": feed.LastChapter,
		"lastChecked": feed.LastChecked,
	})
}

func (s *Server) exportFeeds(c *fiber.Ctx) error {
	feeds, err := s.storage.GetFeeds()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	exportData := fiber.Map{
		"exported": time.Now().Format(time.RFC3339),
		"version":  "1.0",
		"count":    len(feeds),
		"feeds":    feeds,
	}

	c.Set("Content-Disposition", "attachment; filename=shinkan-rebirth-export.json")
	c.Set("Content-Type", "application/json")

	return c.JSON(exportData)
}

func (s *Server) Start(port string) error {
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	log.Printf("🌐 Web UI running at http://localhost%s\n", port)
	return s.app.Listen(port)
}
