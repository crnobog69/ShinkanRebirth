# Shinkan Rebirth - 新刊

A Go application that monitors manga and anime RSS feeds and sends notifications via Gotify when new chapters or episodes are released.

## ✨ Features

**Core Features:**
- **Manga Monitoring**: Track manga chapters from RSS feeds
- **Anime Monitoring**: Track anime episodes with search filtering (e.g., "Dragon Raja" from Nyaa)
- **Dual Notifications**: Send notifications to Discord and/or Gotify
- **Web Interface**: Modern, responsive web UI for managing feeds
- **Statistics Dashboard**: Real-time stats and uptime tracking
- **Search & Filter**: Search feeds and filter by category
- **Import/Export**: Backup and restore your feed lists
- **Automatic Checks**: Scheduled checks via cron expressions
- **Error Handling**: Automatic retry with exponential backoff

**Differences from Original:**
- Written in Go for better performance and lower resource usage
- Supports both Discord AND Gotify notifications (use one or both)
- Separate data files for manga (`mangas.json`) and anime (`anime.json`)
- Anime feeds can search for specific titles in the feed (e.g., "Dragon Raja" from Nyaa)

## 🔄 Migrating from Node.js Version

If you're migrating from the original Node.js Shinkan bot:

1. Copy your `data/mangas.json` to `ShinkanRebirth/data/`
2. The file structure is compatible! Just ensure:
   - `mangas.json` contains your manga feeds
   - `anime.json` contains your anime feeds (with `searchText` field)
3. Each feed needs a `type` field:
   - `"type": "manga"` for manga feeds
   - `"type": "anime"` for anime feeds

The system will automatically read from both files and merge them in the web UI!

## 🚀 Setup

### Prerequisites

- Go 1.21 or higher
- At least one of:
  - A Discord bot token and channel ID, OR
  - A Gotify server instance and app token
- Both Discord and Gotify can be used simultaneously

### 1. Clone and Navigate

```bash
cd ShinkanRebirth
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Edit `.env`:

```env
# Gotify Configuration (optional if Discord is configured)
GOTIFY_SERVER=https://gotify.example.com
GOTIFY_TOKEN=your_gotify_app_token_here

# Discord Configuration (optional if Gotify is configured)
DISCORD_TOKEN=your_discord_bot_token_here
DISCORD_CHANNEL_ID=your_discord_channel_id_here

# Note: You can use both Gotify and Discord, or just one of them

# Web Server Configuration
WEB_PORT=11111

# Check Interval (cron format)
CHECK_INTERVAL=0 * * * *

# Data Storage (separate files for manga and anime)
MANGA_DATA_FILE=./data/mangas.json
ANIME_DATA_FILE=./data/anime.json
```

### 4. Build

```bash
go build -o shinkan-rebirth ./cmd/shinkan
```

### 5. Run

```bash
./shinkan-rebirth
```

### 6. Access Web UI

(You will need to set this one up your self, You can use Caddy.)

Open your browser and navigate to:
```
http://shinkan.local:11111
```

## 📝 Usage

### Adding Manga Feeds

1. Select "📖 Manga" from the feed type dropdown
2. Enter the manga name
3. Enter the RSS feed URL (auto-appends `/rss` if needed)
4. Optionally add AniList URL and category
5. Click "Add Feed"

### Adding Anime Feeds

1. Select "🎬 Anime" from the feed type dropdown
2. Enter the anime name
3. Enter the RSS feed URL (e.g., Nyaa)
4. Enter search text to filter specific titles (e.g., "Dragon Raja")
5. Optionally add AniList URL and category
6. Click "Add Feed"

The anime search feature will look through all items in the RSS feed and only notify when an item matching your search text is found.

### Check Intervals

The check interval uses cron format. Examples:

- `0 * * * *` - Every hour at minute 0 (default)
- `*/30 * * * *` - Every 30 minutes
- `0 */2 * * *` - Every 2 hours
- `0 0 * * *` - Once a day at midnight
- `0 9,21 * * *` - Twice a day (9 AM and 9 PM)

## 🔧 API Endpoints

### Feeds
- `GET /api/feeds` - Get all feeds (supports `?search=` and `?category=` params)
- `POST /api/feeds` - Add new feed
- `PUT /api/feeds/:id` - Update feed
- `DELETE /api/feeds/:id` - Delete feed
- `POST /api/feeds/:id/test` - Send test notification
- `POST /api/feeds/:id/check` - Manually check feed

### Data Management
- `GET /api/export` - Export feeds as JSON
- `POST /api/import` - Import feeds from JSON

### Statistics
- `GET /api/stats` - Get runtime statistics
- `GET /api/categories` - Get all categories
- `GET /api/health` - Health check endpoint

## 💬 Discord Slash Commands

The bot supports Discord slash commands:

- `/check` - Manually trigger a check for all feeds
- `/quote` - Get a random quote from Kafka and others
- `/stats` - Get a link to the web UI statistics

To use slash commands, ensure your Discord bot has the `applications.commands` scope enabled.

## 🎲 Quotes System

Add your own quotes to `data/quotes.json`:

```json
{
  "quotes": [
    "Your custom quote here",
    "Another inspiring quote"
  ]
}
```

The bot comes with quotes from Franz Kafka and other philosophical thoughts.

## 🐳 Running with Docker

### Using docker-compose (Recommended)

1. Copy `.env.docker` to `.env` and configure:
   ```bash
   cp .env.docker .env
   nano .env
   ```

2. Start with docker-compose:
   ```bash
   docker-compose up -d
   ```

3. View logs:
   ```bash
   docker-compose logs -f
   ```

4. Stop:
   ```bash
   docker-compose down
   ```

### Using Docker directly

Build and run:

```bash
# Build
./docker-build.sh

# Or manually:
docker build -t shinkan-rebirth:latest .

# Run
docker run -d \
  --name shinkan-rebirth \
  -p 11111:11111 \
  -v $(pwd)/data:/root/data \
  -e DISCORD_TOKEN=your_token \
  -e DISCORD_CHANNEL_ID=your_channel_id \
  -e GOTIFY_SERVER=https://gotify.example.com \
  -e GOTIFY_TOKEN=your_token \
  shinkan-rebirth:latest
```

### Docker Features

- Persistent data storage via volume mounts
- Timezone configuration
- Hot-reload data files (edit `data/mangas.json` or `data/anime.json` externally)
- Easy updates: rebuild and restart
- Resource efficient Alpine Linux base

## 🔨 Running as a Service

### Systemd Service (Linux)

Create `/etc/systemd/system/shinkan-rebirth.service`:

```ini
[Unit]
Description=Shinkan Rebirth - Manga/Anime RSS Monitor
After=network.target

[Service]
Type=simple
User=your_user
WorkingDirectory=/path/to/ShinkanRebirth
ExecStart=/path/to/ShinkanRebirth/shinkan-rebirth
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable shinkan-rebirth
sudo systemctl start shinkan-rebirth
sudo systemctl status shinkan-rebirth
```

## 🌟 Example RSS Feeds

### Manga
- MangaDex series: `https://mangadex.org/title/[series-id]/rss`

### Anime
- Nyaa user RSS: `https://nyaa.si/?page=rss&u=Erai-raws`
  - Use search text: "Dragon Raja" to filter specific anime

## 🔐 Security Notes

- Keep your `.env` file secure and never commit it to version control
- Use HTTPS for your Gotify server
- Consider running behind a reverse proxy (nginx, Caddy) for the web UI
- Restrict web UI access to trusted networks only

## 🤝 Contributing

This is a rewrite of the original Node.js Shinkan bot in Go. Feel free to:
- Report bugs
- Suggest features
- Submit pull requests

## 📄 License

Same as the original Shinkan project.

## 👤 Author

**crnobog**

Original Node.js version: [Shinkan](https://github.com/crnobog69/shinkan)
Go rewrite: Shinkan Rebirth

---

Made with ♥  and Go
