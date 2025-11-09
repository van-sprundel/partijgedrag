#!/bin/bash
set -e

echo "🚀 Starting zero-downtime deployment..."

# Load environment variables
if [ -f .env.prod ]; then
  export $(cat .env.prod | grep -v '^#' | xargs)
fi

# Pull latest images
echo "📦 Pulling latest Docker images..."
docker compose -f docker-compose.prod.yml pull

# Deploy web service with zero downtime
echo "🌐 Updating web service..."
docker compose -f docker-compose.prod.yml up -d --no-deps --build web

# Wait for health check
echo "⏳ Waiting for web service to be healthy..."
timeout 60 bash -c 'until docker inspect --format="{{.State.Health.Status}}" partijgedrag-web | grep -q "healthy"; do sleep 2; done' || {
  echo "❌ Web service failed health check, rolling back..."
  docker compose -f docker-compose.prod.yml up -d --no-deps --force-recreate web
  exit 1
}

echo "✅ Web service updated successfully!"

# Update ETL (downtime is OK)
echo "📊 Updating ETL service..."
docker compose -f docker-compose.prod.yml up -d --no-deps etl

# Cleanup old images
echo "🧹 Cleaning up old images..."
docker image prune -f

echo "✅ Deployment complete!"
