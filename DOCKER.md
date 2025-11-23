# Docker Quick Reference

## Build

```bash
./docker-build.sh
```

Or manually:
```bash
docker build -t shinkan-rebirth:latest .
```

## Run with docker-compose

```bash
# Start
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down

# Restart
docker-compose restart

# Update (rebuild and restart)
docker-compose down
docker-compose build
docker-compose up -d
```

## Run with docker

```bash
docker run -d \
  --name shinkan-rebirth \
  --restart unless-stopped \
  -p 11111:11111 \
  -v $(pwd)/data:/root/data \
  -e DISCORD_TOKEN=your_token_here \
  -e DISCORD_CHANNEL_ID=your_channel_id_here \
  -e CHECK_INTERVAL="0 * * * *" \
  shinkan-rebirth:latest
```

## Useful Commands

```bash
# View logs
docker logs -f shinkan-rebirth

# Stop
docker stop shinkan-rebirth

# Start
docker start shinkan-rebirth

# Remove
docker rm -f shinkan-rebirth

# Execute command inside container
docker exec -it shinkan-rebirth sh

# Update data files
# Just edit data/mangas.json or data/anime.json on host
# The container will pick up changes automatically
```

## Environment Variables

All environment variables from `.env` can be passed with `-e`:

- `GOTIFY_SERVER` - Gotify server URL
- `GOTIFY_TOKEN` - Gotify app token
- `DISCORD_TOKEN` - Discord bot token
- `DISCORD_CHANNEL_ID` - Discord channel ID
- `WEB_PORT` - Web UI port (default: 11111)
- `CHECK_INTERVAL` - Cron expression for checks
- `TZ` - Timezone (default: Europe/Belgrade)

## Volumes

The container uses these volumes:

- `./data:/root/data` - Persistent data storage

Your manga and anime feeds are stored here and persist across container restarts.

## Troubleshooting

### Container won't start
```bash
docker logs shinkan-rebirth
```

### Check if running
```bash
docker ps | grep shinkan-rebirth
```

### Rebuild after code changes
```bash
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```
