package checker

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"shinkan-rebirth/internal/models"
	"shinkan-rebirth/internal/notifier"
	"shinkan-rebirth/internal/storage"

	"github.com/mmcdole/gofeed"
)

type Checker struct {
	storage  *storage.Storage
	notifier *notifier.Notifier
	parser   *gofeed.Parser
	stats    Stats
	mu       sync.RWMutex
}

type Stats struct {
	TotalChecks       int
	SuccessfulChecks  int
	FailedChecks      int
	NotificationsSent int
	LastCheckTime     *string
}

func New(storage *storage.Storage, notifier *notifier.Notifier) *Checker {
	return &Checker{
		storage:  storage,
		notifier: notifier,
		parser:   gofeed.NewParser(),
		stats:    Stats{},
	}
}

func (c *Checker) GetStats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

func (c *Checker) ResetStats() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats = Stats{}
}

func (c *Checker) CheckFeed(feed models.Feed, retries int) error {
	c.mu.Lock()
	c.stats.TotalChecks++
	c.mu.Unlock()

	var lastErr error
	for attempt := 1; attempt <= retries; attempt++ {
		err := c.checkFeedOnce(feed)
		if err == nil {
			c.mu.Lock()
			c.stats.SuccessfulChecks++
			c.mu.Unlock()
			return nil
		}

		lastErr = err
		log.Printf("✗ [%s] Error (attempt %d/%d): %v\n", feed.Name, attempt, retries, err)

		if attempt < retries {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	c.mu.Lock()
	c.stats.FailedChecks++
	c.mu.Unlock()

	// Update feed with error
	failCount := feed.FailCount + 1
	lastChecked := time.Now().Format(time.RFC3339)
	errorMsg := lastErr.Error()
	c.storage.UpdateFeed(feed.ID, map[string]interface{}{
		"lastChecked": lastChecked,
		"lastError":   errorMsg,
		"failCount":   failCount,
	})

	return lastErr
}

func (c *Checker) checkFeedOnce(feed models.Feed) error {
	rssFeed, err := c.parser.ParseURL(feed.RSSUrl)
	if err != nil {
		return fmt.Errorf("failed to parse RSS: %w", err)
	}

	if len(rssFeed.Items) == 0 {
		return fmt.Errorf("no items found in RSS feed")
	}

	var latestItem *gofeed.Item
	
	// For anime feeds with search text, find matching item
	if feed.Type == models.FeedTypeAnime && feed.SearchText != nil && *feed.SearchText != "" {
		searchText := strings.ToLower(*feed.SearchText)
		for _, item := range rssFeed.Items {
			if strings.Contains(strings.ToLower(item.Title), searchText) {
				latestItem = item
				break
			}
		}
		
		if latestItem == nil {
			log.Printf("✓ [%s] No matching anime found for search: %s\n", feed.Name, *feed.SearchText)
			// Update last checked time even if no match found
			lastChecked := time.Now().Format(time.RFC3339)
			c.storage.UpdateFeed(feed.ID, map[string]interface{}{
				"lastChecked": lastChecked,
				"lastError":   nil,
				"failCount":   0,
			})
			return nil
		}
	} else {
		// For manga or anime without search, just get the first item
		latestItem = rssFeed.Items[0]
	}

	latestChapter := latestItem.Title
	link := latestItem.Link

	// Check if this is a new chapter
	if feed.LastChapter == nil {
		log.Printf("✓ [%s] First check - storing: %s\n", feed.Name, latestChapter)
	} else if *feed.LastChapter != latestChapter {
		log.Printf("🆕 [%s] NEW %s FOUND!\n", feed.Name, 
			map[bool]string{true: "EPISODE", false: "CHAPTER"}[feed.Type == models.FeedTypeAnime])
		log.Printf("   Old: %s\n", *feed.LastChapter)
		log.Printf("   New: %s\n", latestChapter)

		// Send notification
		err := c.notifier.SendNotification(feed.Name, latestChapter, link, string(feed.Type), feed.AnilistUrl, feed.Cover)
		if err != nil {
			log.Printf("⚠️ [%s] Failed to send notification: %v\n", feed.Name, err)
		} else {
			c.mu.Lock()
			c.stats.NotificationsSent++
			c.mu.Unlock()
		}
	} else {
		log.Printf("✓ [%s] No new %s (still: %s)\n", feed.Name,
			map[bool]string{true: "episode", false: "chapter"}[feed.Type == models.FeedTypeAnime],
			latestChapter)
	}

	// Update feed
	lastChecked := time.Now().Format(time.RFC3339)
	c.storage.UpdateFeed(feed.ID, map[string]interface{}{
		"lastChecked": lastChecked,
		"lastChapter": latestChapter,
		"lastError":   nil,
		"failCount":   0,
	})

	return nil
}

func (c *Checker) CheckAll() {
	feeds, err := c.storage.GetFeeds()
	if err != nil {
		log.Printf("❌ Error getting feeds: %v\n", err)
		return
	}

	log.Println(strings.Repeat("=", 50))
	log.Printf("🔍 Checking %d feed(s) at %s\n", len(feeds), time.Now().Format(time.RFC3339))
	log.Println(strings.Repeat("=", 50))

	lastCheckTime := time.Now().Format(time.RFC3339)
	c.mu.Lock()
	c.stats.LastCheckTime = &lastCheckTime
	c.mu.Unlock()

	for _, feed := range feeds {
		c.CheckFeed(feed, 3)
		time.Sleep(500 * time.Millisecond) // Small delay between checks
	}

	c.mu.RLock()
	log.Println(strings.Repeat("=", 50))
	log.Printf("📊 Check complete - Success: %d/%d, Notifications: %d\n",
		c.stats.SuccessfulChecks, c.stats.TotalChecks, c.stats.NotificationsSent)
	log.Println(strings.Repeat("=", 50))
	c.mu.RUnlock()
}

func (c *Checker) TestFeed(feedID string) (map[string]interface{}, error) {
	feeds, err := c.storage.GetFeeds()
	if err != nil {
		return nil, err
	}

	var feed *models.Feed
	for _, f := range feeds {
		if f.ID == feedID {
			feed = &f
			break
		}
	}

	if feed == nil {
		return nil, fmt.Errorf("feed not found")
	}

	rssFeed, err := c.parser.ParseURL(feed.RSSUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS: %w", err)
	}

	if len(rssFeed.Items) == 0 {
		return map[string]interface{}{"error": "No items found in RSS feed"}, nil
	}

	var latestItem *gofeed.Item
	
	// For anime feeds with search text, find matching item
	if feed.Type == models.FeedTypeAnime && feed.SearchText != nil && *feed.SearchText != "" {
		searchText := strings.ToLower(*feed.SearchText)
		for _, item := range rssFeed.Items {
			if strings.Contains(strings.ToLower(item.Title), searchText) {
				latestItem = item
				break
			}
		}
		
		if latestItem == nil {
			return map[string]interface{}{
				"error": fmt.Sprintf("No matching items found for search: %s", *feed.SearchText),
			}, nil
		}
	} else {
		latestItem = rssFeed.Items[0]
	}

	// Send test notification
	err = c.notifier.SendTestNotification(feed.Name, latestItem.Title, latestItem.Link, string(feed.Type), feed.AnilistUrl, feed.Cover)
	if err != nil {
		return nil, fmt.Errorf("failed to send test notification: %w", err)
	}

	result := map[string]interface{}{
		"title": latestItem.Title,
		"link":  latestItem.Link,
		"sent":  true,
	}

	if latestItem.PublishedParsed != nil {
		result["date"] = latestItem.PublishedParsed.Format(time.RFC3339)
	}

	return result, nil
}
