#!/bin/bash
set -e

echo "🚀 Production Deployment (Docker Swarm)"
echo "========================================="
echo ""

# Load environment
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
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

# Check if Swarm is initialized
if ! docker info 2>/dev/null | grep -q "Swarm: active"; then
  echo "⚠️  Docker Swarm is not initialized"
  echo ""
  echo "Docker Swarm provides PROPER zero-downtime deployments with:"
  echo "  • Automatic rolling updates"
  echo "  • Health check integration"
  echo "  • Start-first update strategy (new container before stopping old)"
  echo "  • Automatic rollback on failure"
  echo ""
  read -p "Initialize Docker Swarm now? (y/n) " -n 1 -r
  echo ""

  if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🔧 Initializing Docker Swarm..."
    docker swarm init
    echo "✅ Swarm initialized!"
    echo ""
  else
    echo "❌ Swarm required for zero-downtime deployments"
    echo "   Alternative: Use deploy-compose.sh (may have brief downtime)"
    exit 1
  fi
fi

# Pull images
echo "📥 Pulling images..."
docker pull "$WEB_IMAGE"
docker pull "$ETL_IMAGE"
echo ""

# Deploy stack with rolling update
echo "🚀 Deploying stack with zero-downtime rolling update..."
docker stack deploy -c docker-stack.yml --with-registry-auth partijgedrag

# Monitor deployment
echo ""
echo "⏳ Monitoring deployment progress..."
echo ""

timeout 120 bash -c '
  while true; do
    REPLICAS=$(docker service ls --filter "name=partijgedrag_web" --format "{{.Replicas}}")
    echo "   Web service: $REPLICAS"

    # Check if service has converged (e.g., "2/2")
    CURRENT=$(echo "$REPLICAS" | cut -d "/" -f 1)
    DESIRED=$(echo "$REPLICAS" | cut -d "/" -f 2)

    if [ "$CURRENT" = "$DESIRED" ] && [ "$CURRENT" -gt "0" ]; then
      echo ""
      echo "✅ Service has converged!"
      break
    fi

    sleep 5
  done
' || {
  echo ""
  echo "⚠️  Deployment is taking longer than expected"
  echo "   Check status with: docker service ps partijgedrag_web"
  exit 1
}

# Verify health
echo ""
echo "🔍 Verifying deployment..."
sleep 10

# Test health endpoint
if curl -f "http://localhost:3000/ready" > /dev/null 2>&1; then
  echo "✅ Health check passed!"
else
  echo "⚠️  Health check warning (service might not be exposed on localhost:3000)"
fi

# Show final state
echo ""
echo "📊 Deployment status:"
docker stack services partijgedrag
echo ""
docker service ps partijgedrag_web --no-trunc

# Cleanup
echo ""
echo "🧹 Cleaning up old images..."
docker image prune -f > /dev/null 2>&1

echo ""
echo "✅ Deployment complete!"
echo "   Version: ${GIT_SHA:-$WEB_IMAGE}"
echo ""
echo "📝 Useful commands:"
echo "   View logs:     docker service logs partijgedrag_web -f"
echo "   Scale service: docker service scale partijgedrag_web=3"
echo "   Rollback:      docker service update --rollback partijgedrag_web"
echo ""
echo "💡 To rollback: Set WEB_IMAGE to previous SHA in .env and re-run"
