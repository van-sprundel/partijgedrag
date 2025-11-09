#!/bin/bash
set -e

echo "🚀 Blue-Green Zero-Downtime Deployment"
echo "======================================="
echo ""

# Configuration
WEB_IMAGE=${WEB_IMAGE:-ghcr.io/van-sprundel/partijgedrag-web:latest}
ETL_IMAGE=${ETL_IMAGE:-ghcr.io/van-sprundel/partijgedrag-etl:latest}
HEALTH_ENDPOINT="http://localhost:8080/ready"
NETWORK="partijgedrag"

# Load environment
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

echo "📦 Target images:"
echo "   Web: $WEB_IMAGE"
echo "   ETL: $ETL_IMAGE"
echo ""

# Pull images
echo "📥 Pulling images..."
docker pull "$WEB_IMAGE"
docker pull "$ETL_IMAGE"

# Create network if it doesn't exist
docker network create "$NETWORK" 2>/dev/null || true

# Determine which color is currently active
if docker ps --filter "name=partijgedrag-app-blue" --format "{{.Names}}" | grep -q "blue"; then
  ACTIVE="blue"
  INACTIVE="green"
else
  ACTIVE="green"
  INACTIVE="blue"
fi

echo ""
echo "📊 Current active: $ACTIVE"
echo "   Deploying to: $INACTIVE"
echo ""

# Start new container (inactive color)
echo "🚀 Starting $INACTIVE container..."
docker run -d \
  --name "partijgedrag-app-$INACTIVE" \
  --network "$NETWORK" \
  -p "808$([ "$INACTIVE" = "blue" ] && echo "0" || echo "1"):80" \
  -e DATABASE_URL="$DATABASE_URL" \
  -e PORT=80 \
  -e NODE_ENV=production \
  -e CORS_ORIGIN="$CORS_ORIGIN" \
  --health-cmd="wget --quiet --tries=1 --spider http://localhost:80/ready || exit 1" \
  --health-interval=10s \
  --health-timeout=5s \
  --health-retries=3 \
  --health-start-period=40s \
  --restart unless-stopped \
  "$WEB_IMAGE"

# Wait for health check
echo "⏳ Waiting for $INACTIVE to be healthy..."
TIMEOUT=60
ELAPSED=0
while [ $ELAPSED -lt $TIMEOUT ]; do
  HEALTH=$(docker inspect --format='{{.State.Health.Status}}' "partijgedrag-app-$INACTIVE" 2>/dev/null || echo "starting")

  if [ "$HEALTH" = "healthy" ]; then
    echo "✅ $INACTIVE is healthy!"
    break
  fi

  echo "   Status: $HEALTH (${ELAPSED}s / ${TIMEOUT}s)"
  sleep 5
  ELAPSED=$((ELAPSED + 5))
done

if [ $ELAPSED -ge $TIMEOUT ]; then
  echo "❌ $INACTIVE failed to become healthy within ${TIMEOUT}s"
  echo ""
  echo "Container logs:"
  docker logs --tail 50 "partijgedrag-app-$INACTIVE"
  docker stop "partijgedrag-app-$INACTIVE"
  docker rm "partijgedrag-app-$INACTIVE"
  exit 1
fi

# Test the endpoint directly
echo ""
echo "🔍 Testing $INACTIVE endpoint..."
INACTIVE_PORT=$([ "$INACTIVE" = "blue" ] && echo "8080" || echo "8081")
if curl -f "http://localhost:$INACTIVE_PORT/ready" > /dev/null 2>&1; then
  echo "✅ Endpoint test passed!"
else
  echo "❌ Endpoint test failed"
  docker logs --tail 50 "partijgedrag-app-$INACTIVE"
  docker stop "partijgedrag-app-$INACTIVE"
  docker rm "partijgedrag-app-$INACTIVE"
  exit 1
fi

# Switch traffic: Update Nginx upstream or just inform user
echo ""
echo "✅ $INACTIVE is ready to receive traffic!"
echo ""
echo "🔄 Now you should update your Nginx config to point to port $INACTIVE_PORT"
echo "   Or switch the ports in your load balancer"
echo ""
read -p "Press Enter after you've switched traffic to $INACTIVE..."

# Stop old container
if docker ps -a --filter "name=partijgedrag-app-$ACTIVE" --format "{{.Names}}" | grep -q "$ACTIVE"; then
  echo ""
  echo "🛑 Stopping old $ACTIVE container..."
  docker stop "partijgedrag-app-$ACTIVE" --time 30
  docker rm "partijgedrag-app-$ACTIVE"
  echo "✅ Old container removed"
fi

# Update ETL
echo ""
echo "📊 Updating ETL..."
docker stop partijgedrag-etl 2>/dev/null || true
docker rm partijgedrag-etl 2>/dev/null || true
docker run -d \
  --name partijgedrag-etl \
  --network "$NETWORK" \
  -e DATABASE_URL="$DATABASE_URL" \
  --restart unless-stopped \
  "$ETL_IMAGE"

echo ""
echo "✅ Deployment complete!"
echo ""
echo "   Active container: partijgedrag-app-$INACTIVE"
echo "   Port: $INACTIVE_PORT"
echo "   Version: ${GIT_SHA:-$WEB_IMAGE}"
echo ""
echo "To rollback: Set WEB_IMAGE to previous version and re-run this script"
