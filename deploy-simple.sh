#!/bin/bash
set -e

# Load environment variables from .env file
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

echo "🚀 Zero-Downtime Deployment (Simple Method)"
echo ""

COMPOSE_FILE="docker-compose.server.yml"

# Pull latest images
echo "📦 Pulling latest images..."
docker compose -f $COMPOSE_FILE pull

echo ""
echo "🔄 Deploying with zero downtime..."

# The trick: Scale to 2, update, then scale back to 1
# Nginx load balancer will distribute traffic between both instances

# Step 1: Start second instance (still running old version)
docker compose -f $COMPOSE_FILE up -d --scale app=2 --no-recreate

echo "⏳ Waiting for second instance to start..."
docker compose -f $COMPOSE_FILE up -d --scale app=2 --no-recreate --wait

# Step 2: Now recreate both (Docker will rolling restart them one at a time)
echo "⏳ Waiting for new instances to be healthy..."
docker compose -f $COMPOSE_FILE up -d --scale app=2 --force-recreate --wait app

# Step 3: Verify both are running
echo "📊 Checking status..."
docker compose -f $COMPOSE_FILE ps app

# Step 4: (Optional) Scale back to 1 instance
if [[ "$1" == "--scale-down" ]]; then
    docker compose -f $COMPOSE_FILE up -d --scale app=1 --no-recreate
    echo "📉 Scaled back to 1 instance"
else
    echo "✅ Keeping 2 instances for redundancy. Pass --scale-down to scale back."
fi

# Update ETL (downtime OK)
echo ""
echo "🔄 Updating ETL..."
docker compose -f $COMPOSE_FILE up -d --force-recreate etl

echo ""
echo "✅ Deployment complete!"
docker compose -f $COMPOSE_FILE ps
