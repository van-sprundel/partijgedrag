#!/bin/bash
set -e

echo "🚀 Production Deployment (Docker Compose)"
echo "=========================================="
echo ""
echo "⚠️  WARNING: Docker Compose does NOT guarantee zero-downtime"
echo "    For true zero-downtime, use deploy-production.sh (Docker Swarm)"
echo ""

# Load environment
if [ -f .env ]; then
  set -a
  # shellcheck source=.env
  source .env
  set +a
fi

# Validate
if [ -z "$WEB_IMAGE" ] || [ -z "$DATABASE_URL" ]; then
  echo "❌ Error: WEB_IMAGE and DATABASE_URL must be set in .env"
  exit 1
fi

ETL_IMAGE=${ETL_IMAGE:-ghcr.io/van-sprundel/partijgedrag-etl:latest}

echo "📦 Deploying version: ${GIT_SHA:-unknown}"
echo "   Web: $WEB_IMAGE"
echo "   ETL: $ETL_IMAGE"
echo ""

# Check Docker Compose version
COMPOSE_VERSION=$(docker compose version --short 2>/dev/null || echo "unknown")
echo "📋 Docker Compose version: $COMPOSE_VERSION"

REQUIRED_COMPOSE_VERSION="2.20"
if [ "$(printf '%s\n' "$REQUIRED_COMPOSE_VERSION" "$COMPOSE_VERSION" | sort -V | head -n1)" != "$REQUIRED_COMPOSE_VERSION" ] && [ "$COMPOSE_VERSION" != "unknown" ]; then
  echo "⚠️  Docker Compose v${REQUIRED_COMPOSE_VERSION}+ recommended for --wait flag"
fi
echo ""

# Pull images
echo "📥 Pulling images..."
docker compose -f docker-compose.server.yml pull

echo ""
echo "🔄 Deploying..."
echo ""
echo "This deployment strategy:"
echo "  1. Starts new containers with new image"
echo "  2. Waits for health checks (if Docker Compose v2.20+)"
echo "  3. Stops old containers"
echo ""
echo "⚠️  There may be a brief moment (1-5s) where no containers are running"
echo "    during the transition between old and new containers."
echo ""

read -p "Continue with deployment? (y/n) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "❌ Deployment cancelled"
  exit 1
fi

# Deploy with wait flag (best effort)
echo ""
echo "🚀 Starting deployment..."

if [ "$(printf '%s\n' "$REQUIRED_COMPOSE_VERSION" "$COMPOSE_VERSION" | sort -V | head -n1)" = "$REQUIRED_COMPOSE_VERSION" ] && [ "$COMPOSE_VERSION" != "unknown" ]; then
  # Use --wait for Docker Compose 2.20+
  docker compose -f docker-compose.server.yml up -d --wait
else
  # Fallback for older versions
  docker compose -f docker-compose.server.yml up -d
  echo "⏳ Waiting 30s for containers to stabilize..."
  sleep 30
fi

# Verify deployment
echo ""
echo "🔍 Verifying deployment..."
sleep 5

if curl -f "http://localhost:8080/ready" > /dev/null 2>&1; then
  echo "✅ Health check passed!"
else
  echo "❌ Health check failed!"
  echo ""
  echo "Container logs:"
  docker compose -f docker-compose.server.yml logs --tail 50 app
  exit 1
fi

# Show status
echo ""
echo "📊 Deployment status:"
docker compose -f docker-compose.server.yml ps

# Cleanup
echo ""
echo "🧹 Cleaning up old images..."
docker image prune -f

echo ""
echo "✅ Deployment complete!"
echo "   Version: ${GIT_SHA:-$WEB_IMAGE}"
echo ""
echo "💡 For zero-downtime deployments, use: ./deploy-production.sh (Docker Swarm)"
