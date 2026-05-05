#!/bin/bash

echo "🚀 Connecting to server and pulling latest code..."
ssh root@168.144.116.143 "cd ~/customurls && git pull origin main"

echo "🐳 Rebuilding Docker containers..."
ssh root@168.144.116.143 "cd ~/customurls && docker-compose down && docker-compose up -d --build"

echo "🧹 Cleaning unused Docker resources..."
ssh root@168.144.116.143 "docker system prune -f"

echo "🔄 Checking container status..."
ssh root@168.144.116.143 "docker ps"

echo "🌍 Deployment complete! Your app is live at https://customurls.in"
