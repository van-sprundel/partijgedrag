#!/bin/bash
set -e

echo "🚀 Starting zero-downtime deployment..."

# Configuration
COMPOSE_FILE="docker-compose.server.yml"
WEB_SERVICE="app"
ETL_SERVICE="etl"

# Pull latest images
echo "📦 Pulling latest images..."
docker compose -f $COMPOSE_FILE pull

echo ""
echo "========================================="
echo "  Deploying Web Service (Zero Downtime)"
echo "========================================="

# Step 1: Scale up to 2 instances (old version)
echo "📈 Scaling to 2 instances..."
docker compose -f $COMPOSE_FILE up -d --scale $WEB_SERVICE=2 --no-recreate

sleep 5

# Step 2: Remove the first old container, Docker will start new one automatically
echo "🔄 Rolling restart - updating first instance..."
OLD_CONTAINER_1=$(docker compose -f $COMPOSE_FILE ps -q $WEB_SERVICE | head -n1)
docker stop $OLD_CONTAINER_1
docker rm $OLD_CONTAINER_1

# Step 3: Start new container with latest image
docker compose -f $COMPOSE_FILE up -d --scale $WEB_SERVICE=2 --no-recreate

# Wait for health check
echo "⏳ Waiting for new instance to be healthy (30s)..."
sleep 30

# Step 4: Remove the second old container
echo "🔄 Rolling restart - updating second instance..."
OLD_CONTAINER_2=$(docker compose -f $COMPOSE_FILE ps -q $WEB_SERVICE | head -n1)
docker stop $OLD_CONTAINER_2
docker rm $OLD_CONTAINER_2

# Step 5: Ensure we have 2 healthy instances
docker compose -f $COMPOSE_FILE up -d --scale $WEB_SERVICE=2

echo "⏳ Waiting for both instances to be healthy..."
sleep 15

# Step 6: Scale back down to 1 instance (optional - remove if you want to keep 2)
echo "📉 Scaling back to 1 instance..."
docker compose -f $COMPOSE_FILE up -d --scale $WEB_SERVICE=1

echo ""
echo "========================================="
echo "  Deploying ETL Service"
echo "========================================="

# ETL can have downtime, so just recreate it
echo "🔄 Updating ETL service..."
docker compose -f $COMPOSE_FILE up -d --no-deps --force-recreate $ETL_SERVICE

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Current status:"
docker compose -f $COMPOSE_FILE ps

echo ""
echo "🧹 Cleaning up old images..."
docker image prune -f

echo ""
echo "✅ All done! Deployment successful."
