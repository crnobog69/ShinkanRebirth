# Quick Setup Guide

## 1. Configure Environment

```bash
cp .env.example .env
nano .env  # or use your preferred editor
```

Set either Discord OR Gotify (or both):

**For Discord:**
```env
DISCORD_TOKEN=your_bot_token_here
DISCORD_CHANNEL_ID=your_channel_id_here
```

**For Gotify:**
```env
GOTIFY_SERVER=https://gotify.example.com
GOTIFY_TOKEN=your_app_token_here
```

## 2. Install Dependencies

```bash
go mod download
```

## 3. Build

```bash
go build -o shinkan-rebirth ./cmd/shinkan
```

## 4. Run

```bash
./shinkan-rebirth
```

## 5. Access Web UI

Open: http://localhost:11111

## Adding Feeds

### Manga
1. Select "📖 Manga"
2. Enter manga name and RSS URL
3. Optional: Add AniList URL and category

### Anime (with search)
1. Select "🎬 Anime"
2. Enter anime name
3. Enter RSS URL (e.g., `https://nyaa.si/?page=rss&u=Erai-raws`)
4. Enter search text (e.g., "Dragon Raja")
5. Optional: Add AniList URL and category

The anime search will look for the text in feed item titles and only notify when a match is found.

## Discord Bot Setup

1. Go to https://discord.com/developers/applications
2. Create a new application
3. Go to "Bot" section and create a bot
4. Copy the bot token
5. Enable "MESSAGE CONTENT INTENT" (optional)
6. Go to OAuth2 > URL Generator
7. Select scopes: `bot`
8. Select permissions: `Send Messages`, `Embed Links`
9. Use the generated URL to invite the bot to your server
10. Right-click your channel > Copy ID (enable Developer Mode in settings)

## Gotify Server Setup

1. Install Gotify server: https://gotify.net/
2. Create an application in Gotify
3. Copy the application token
4. Use your Gotify server URL and token in `.env`

## Using Both Discord and Gotify

Simply configure both in your `.env` file! Notifications will be sent to both services.
