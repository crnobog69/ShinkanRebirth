package models

// FeedType represents the type of feed (manga or anime)
type FeedType string

const (
	FeedTypeManga FeedType = "manga"
	FeedTypeAnime FeedType = "anime"
)

// Feed represents a manga or anime RSS feed to monitor
type Feed struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	RSSUrl      string    `json:"rssUrl"`
	Type        FeedType  `json:"type"` // "manga" or "anime"
	AnilistUrl  *string   `json:"anilistUrl,omitempty"`
	Category    string    `json:"category"`
	LastChecked *string   `json:"lastChecked"`
	LastChapter *string   `json:"lastChapter"`
	LastError   *string   `json:"lastError"`
	FailCount   int       `json:"failCount"`
	AddedAt     string    `json:"addedAt"`
	SearchText  *string   `json:"searchText,omitempty"` // For anime: text to search for (e.g., "Dragon Raja")
	Cover       *string   `json:"cover,omitempty"` // Cover image URL for Discord embeds
}

// Storage represents the data structure for storing feeds
type Storage struct {
	Feeds []Feed `json:"feeds"`
}

// Stats represents runtime statistics
type Stats struct {
	TotalChecks       int       `json:"totalChecks"`
	SuccessfulChecks  int       `json:"successfulChecks"`
	FailedChecks      int       `json:"failedChecks"`
	NotificationsSent int       `json:"notificationsSent"`
	LastCheckTime     *string   `json:"lastCheckTime"`
	TotalFeeds        int       `json:"totalFeeds"`
	FeedsWithErrors   int       `json:"feedsWithErrors"`
	FeedsNeverChecked int       `json:"feedsNeverChecked"`
	Categories        int       `json:"categories"`
	Uptime            int64     `json:"uptime"`
}

// GotifyMessage represents a message to send to Gotify
type GotifyMessage struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
	Extras   map[string]interface{} `json:"extras,omitempty"`
}
