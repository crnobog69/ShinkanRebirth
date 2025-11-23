#!/bin/bash

echo "🐳 Building Shinkan Rebirth Docker image..."
docker build -t shinkan-rebirth:latest .

if [ $? -eq 0 ]; then
    echo "✅ Docker image built successfully!"
    echo ""
    echo "To run with docker-compose:"
    echo "  docker-compose up -d"
    echo ""
    echo "To run with docker:"
    echo "  docker run -d --name shinkan-rebirth \\"
    echo "    -p 11111:11111 \\"
    echo "    -v \$(pwd)/data:/root/data \\"
    echo "    -e DISCORD_TOKEN=your_token \\"
    echo "    -e DISCORD_CHANNEL_ID=your_channel_id \\"
    echo "    shinkan-rebirth:latest"
else
    echo "❌ Docker build failed!"
    exit 1
fi
