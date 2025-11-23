package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"shinkan-rebirth/internal/models"

	"github.com/bwmarrin/discordgo"
)

type Notifier struct {
	gotifyServer     string
	gotifyToken      string
	discordSession   *discordgo.Session
	discordChannelID string
	httpClient       *http.Client
	commandHandlers  map[string]func(*discordgo.Session, *discordgo.InteractionCreate)
}

func New(gotifyServer, gotifyToken, discordToken, discordChannelID string) *Notifier {
	n := &Notifier{
		gotifyServer:     strings.TrimSuffix(gotifyServer, "/"),
		gotifyToken:      gotifyToken,
		discordChannelID: discordChannelID,
		httpClient:       &http.Client{},
		commandHandlers:  make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate)),
	}

	// Initialize Discord if token provided
	if discordToken != "" {
		session, err := discordgo.New("Bot " + discordToken)
		if err != nil {
			log.Printf("⚠️ Failed to create Discord session: %v\n", err)
		} else {
			session.Identify.Intents = discordgo.IntentsGuilds
			err = session.Open()
			if err != nil {
				log.Printf("⚠️ Failed to open Discord connection: %v\n", err)
			} else {
				n.discordSession = session
				log.Println("✅ Discord bot connected")
				
				// Set bot presence
				session.UpdateStatusComplex(discordgo.UpdateStatusData{
					Activities: []*discordgo.Activity{{
						Name: "manga & anime",
						Type: discordgo.ActivityTypeWatching,
					}},
					Status: "dnd",
				})
			}
		}
	}

	return n
}

func (n *Notifier) SendNotification(feedName, chapter, link, feedType string, anilistUrl *string) error {
	var title, message string
	var priority int
	var color int

	if feedType == string(models.FeedTypeAnime) {
		title = "🎬 New Anime Episode!"
		message = fmt.Sprintf("**%s**\n%s", feedName, chapter)
		priority = 7
		color = 0x89b4fa // Blue for anime
	} else {
		title = "📖 New Manga Chapter!"
		message = fmt.Sprintf("**%s**\n%s", feedName, chapter)
		priority = 5
		color = 0xa6e3a1 // Green for manga
	}

	if anilistUrl != nil && *anilistUrl != "" {
		message += fmt.Sprintf("\n\n📺 AniList: %s", *anilistUrl)
	}

	message += fmt.Sprintf("\n\n🔗 Link: %s", link)

	// Send to Gotify if configured
	if n.gotifyServer != "" && n.gotifyToken != "" {
		gotifyMsg := models.GotifyMessage{
			Title:    title,
			Message:  message,
			Priority: priority,
			Extras: map[string]interface{}{
				"client::display": map[string]interface{}{
					"contentType": "text/markdown",
				},
			},
		}
		if err := n.sendToGotify(gotifyMsg); err != nil {
			log.Printf("⚠️ Gotify notification failed: %v\n", err)
		}
	}

	// Send to Discord if configured
	if n.discordSession != nil && n.discordChannelID != "" {
		if err := n.sendToDiscord(title, feedName, chapter, link, anilistUrl, color); err != nil {
			log.Printf("⚠️ Discord notification failed: %v\n", err)
		}
	}

	return nil
}

func (n *Notifier) SendTestNotification(feedName, chapter, link, feedType string, anilistUrl *string) error {
	var title, message string
	var color int

	if feedType == string(models.FeedTypeAnime) {
		title = "🧪 TEST: Anime Notification"
		color = 0x89b4fa
	} else {
		title = "🧪 TEST: Manga Notification"
		color = 0xa6e3a1
	}

	message = fmt.Sprintf("**%s**\n%s", feedName, chapter)

	if anilistUrl != nil && *anilistUrl != "" {
		message += fmt.Sprintf("\n\n📺 AniList: %s", *anilistUrl)
	}

	message += fmt.Sprintf("\n\n🔗 Link: %s", link)

	// Send to Gotify if configured
	if n.gotifyServer != "" && n.gotifyToken != "" {
		gotifyMsg := models.GotifyMessage{
			Title:    title,
			Message:  message,
			Priority: 3,
			Extras: map[string]interface{}{
				"client::display": map[string]interface{}{
					"contentType": "text/markdown",
				},
			},
		}
		if err := n.sendToGotify(gotifyMsg); err != nil {
			log.Printf("⚠️ Gotify test notification failed: %v\n", err)
		}
	}

	// Send to Discord if configured
	if n.discordSession != nil && n.discordChannelID != "" {
		if err := n.sendToDiscord(title, feedName, chapter, link, anilistUrl, color); err != nil {
			log.Printf("⚠️ Discord test notification failed: %v\n", err)
		}
	}

	return nil
}

func (n *Notifier) sendToGotify(msg models.GotifyMessage) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	url := fmt.Sprintf("%s/message?token=%s", n.gotifyServer, n.gotifyToken)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gotify returned status %d", resp.StatusCode)
	}

	return nil
}

func (n *Notifier) sendToDiscord(title, feedName, chapter, link string, anilistUrl *string, color int) error {
	description := fmt.Sprintf("**%s**\n%s", feedName, chapter)
	
	if anilistUrl != nil && *anilistUrl != "" {
		description += fmt.Sprintf("\n\n[📺 View on AniList](%s)", *anilistUrl)
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		URL:         link,
		Color:       color,
		Timestamp:   fmt.Sprintf("%d", int64(0)), // Will be auto-set by Discord
	}

	_, err := n.discordSession.ChannelMessageSendEmbed(n.discordChannelID, embed)
	if err != nil {
		return fmt.Errorf("failed to send Discord message: %w", err)
	}

	return nil
}

func (n *Notifier) RegisterCommands(checkCallback func(), quoteCallback func() string) error {
	if n.discordSession == nil {
		return nil
	}

	// Register slash commands
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "check",
			Description: "Manually check all feeds for new chapters/episodes",
		},
		{
			Name:        "quote",
			Description: "Get a random quote",
		},
		{
			Name:        "stats",
			Description: "Show bot statistics",
		},
	}

	// Register commands with Discord
	for _, cmd := range commands {
		_, err := n.discordSession.ApplicationCommandCreate(n.discordSession.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("⚠️ Cannot create '%s' command: %v", cmd.Name, err)
		}
	}

	// Setup command handlers
	n.discordSession.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.ApplicationCommandData().Name {
		case "check":
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "🔍 Starting manual check...",
				},
			})
			go checkCallback()

		case "quote":
			quote := quoteCallback()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: quote,
				},
			})

		case "stats":
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "📊 Check the web UI at http://shinkan.local:11111 for detailed stats!",
				},
			})
		}
	})

	log.Println("✅ Discord slash commands registered")
	return nil
}

func (n *Notifier) Close() {
	if n.discordSession != nil {
		n.discordSession.Close()
	}
}
