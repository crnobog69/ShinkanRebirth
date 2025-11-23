package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"shinkan-rebirth/internal/models"
)

type Storage struct {
	mangaFilePath string
	animeFilePath string
	mu            sync.RWMutex
}

func New(mangaFilePath, animeFilePath string) *Storage {
	s := &Storage{
		mangaFilePath: mangaFilePath,
		animeFilePath: animeFilePath,
	}
	s.ensureDataFiles()
	return s
}

func (s *Storage) ensureDataFiles() {
	// Ensure manga file
	dir := filepath.Dir(s.mangaFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create data directory: %v", err))
	}

	if _, err := os.Stat(s.mangaFilePath); os.IsNotExist(err) {
		data := models.Storage{Feeds: []models.Feed{}}
		if err := s.writeToFile(s.mangaFilePath, data); err != nil {
			panic(fmt.Sprintf("Failed to create manga data file: %v", err))
		}
	}

	// Ensure anime file
	dir = filepath.Dir(s.animeFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create data directory: %v", err))
	}

	if _, err := os.Stat(s.animeFilePath); os.IsNotExist(err) {
		data := models.Storage{Feeds: []models.Feed{}}
		if err := s.writeToFile(s.animeFilePath, data); err != nil {
			panic(fmt.Sprintf("Failed to create anime data file: %v", err))
		}
	}
}

func (s *Storage) readFromFile(filePath string) (models.Storage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return models.Storage{}, err
	}

	var storage models.Storage
	if err := json.Unmarshal(data, &storage); err != nil {
		return models.Storage{}, err
	}

	return storage, nil
}

func (s *Storage) writeToFile(filePath string, data models.Storage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, jsonData, 0644)
}

func (s *Storage) read() (models.Storage, error) {
	mangaData, err := s.readFromFile(s.mangaFilePath)
	if err != nil {
		return models.Storage{}, err
	}

	animeData, err := s.readFromFile(s.animeFilePath)
	if err != nil {
		return models.Storage{}, err
	}

	// Combine both
	combined := models.Storage{
		Feeds: append(mangaData.Feeds, animeData.Feeds...),
	}

	return combined, nil
}

func (s *Storage) write(data models.Storage) error {
	// Split feeds by type
	mangaFeeds := make([]models.Feed, 0)
	animeFeeds := make([]models.Feed, 0)

	for _, feed := range data.Feeds {
		if feed.Type == models.FeedTypeAnime {
			animeFeeds = append(animeFeeds, feed)
		} else {
			mangaFeeds = append(mangaFeeds, feed)
		}
	}

	// Write to separate files
	if err := s.writeToFile(s.mangaFilePath, models.Storage{Feeds: mangaFeeds}); err != nil {
		return err
	}

	if err := s.writeToFile(s.animeFilePath, models.Storage{Feeds: animeFeeds}); err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetFeeds() ([]models.Feed, error) {
	data, err := s.read()
	if err != nil {
		return nil, err
	}
	return data.Feeds, nil
}

func (s *Storage) AddFeed(feed models.Feed) (models.Feed, error) {
	data, err := s.read()
	if err != nil {
		return models.Feed{}, err
	}

	feed.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	feed.AddedAt = time.Now().Format(time.RFC3339)
	feed.FailCount = 0

	if feed.Category == "" {
		feed.Category = "Uncategorized"
	}

	data.Feeds = append(data.Feeds, feed)

	if err := s.write(data); err != nil {
		return models.Feed{}, err
	}

	return feed, nil
}

func (s *Storage) DeleteFeed(id string) error {
	data, err := s.read()
	if err != nil {
		return err
	}

	newFeeds := make([]models.Feed, 0)
	for _, feed := range data.Feeds {
		if feed.ID != id {
			newFeeds = append(newFeeds, feed)
		}
	}

	data.Feeds = newFeeds
	return s.write(data)
}

func (s *Storage) UpdateFeed(id string, updates map[string]interface{}) (*models.Feed, error) {
	data, err := s.read()
	if err != nil {
		return nil, err
	}

	var updatedFeed *models.Feed
	for i, feed := range data.Feeds {
		if feed.ID == id {
			// Apply updates
			if name, ok := updates["name"].(string); ok {
				data.Feeds[i].Name = name
			}
			if rssUrl, ok := updates["rssUrl"].(string); ok {
				data.Feeds[i].RSSUrl = rssUrl
			}
			if anilistUrl, ok := updates["anilistUrl"].(string); ok {
				data.Feeds[i].AnilistUrl = &anilistUrl
			}
			if category, ok := updates["category"].(string); ok {
				data.Feeds[i].Category = category
			}
			if feedType, ok := updates["type"].(string); ok {
				data.Feeds[i].Type = models.FeedType(feedType)
			}
			if searchText, ok := updates["searchText"].(string); ok {
				data.Feeds[i].SearchText = &searchText
			}
			if lastChecked, ok := updates["lastChecked"].(string); ok {
				data.Feeds[i].LastChecked = &lastChecked
			}
			if lastChapter, ok := updates["lastChapter"].(string); ok {
				data.Feeds[i].LastChapter = &lastChapter
			}
			if lastError, ok := updates["lastError"].(string); ok {
				data.Feeds[i].LastError = &lastError
			}
			if lastError := updates["lastError"]; lastError == nil {
				data.Feeds[i].LastError = nil
			}
			if failCount, ok := updates["failCount"].(int); ok {
				data.Feeds[i].FailCount = failCount
			}

			updatedFeed = &data.Feeds[i]
			break
		}
	}

	if updatedFeed == nil {
		return nil, fmt.Errorf("feed not found")
	}

	if err := s.write(data); err != nil {
		return nil, err
	}

	return updatedFeed, nil
}

func (s *Storage) GetCategories() ([]string, error) {
	data, err := s.read()
	if err != nil {
		return nil, err
	}

	categoryMap := make(map[string]bool)
	for _, feed := range data.Feeds {
		category := feed.Category
		if category == "" {
			category = "Uncategorized"
		}
		categoryMap[category] = true
	}

	categories := make([]string, 0, len(categoryMap))
	for category := range categoryMap {
		categories = append(categories, category)
	}

	return categories, nil
}

func (s *Storage) SearchFeeds(query string) ([]models.Feed, error) {
	data, err := s.read()
	if err != nil {
		return nil, err
	}

	// Simple case-insensitive search
	results := make([]models.Feed, 0)
	for _, feed := range data.Feeds {
		if contains(feed.Name, query) ||
			contains(feed.RSSUrl, query) ||
			(feed.LastChapter != nil && contains(*feed.LastChapter, query)) {
			results = append(results, feed)
		}
	}

	return results, nil
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		(s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	// Simple case-insensitive substring search
	sLower := toLower(s)
	substrLower := toLower(substr)
	
	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

func (s *Storage) ImportFeeds(feeds []models.Feed) (int, int, error) {
	data, err := s.read()
	if err != nil {
		return 0, 0, err
	}

	imported := 0
	skipped := 0

	// Create a map of existing RSS URLs for faster lookup
	existingUrls := make(map[string]bool)
	for _, feed := range data.Feeds {
		existingUrls[feed.RSSUrl] = true
	}

	for _, feed := range feeds {
		if existingUrls[feed.RSSUrl] {
			skipped++
			continue
		}

		feed.ID = fmt.Sprintf("%d%d", time.Now().UnixNano(), imported)
		feed.AddedAt = time.Now().Format(time.RFC3339)
		feed.FailCount = 0
		feed.LastChecked = nil
		feed.LastChapter = nil
		feed.LastError = nil

		if feed.Category == "" {
			feed.Category = "Uncategorized"
		}

		data.Feeds = append(data.Feeds, feed)
		imported++
	}

	if err := s.write(data); err != nil {
		return 0, 0, err
	}

	return imported, skipped, nil
}
