#!/bin/bash
set -e

echo "🚀 Starting Docker Swarm deployment..."

# Load environment variables
# Load environment variables
if [ -f .env ]; then
  set -a
  # shellcheck source=.env
  source .env
  set +a
fi

# Pull images from .env
echo "📦 Pulling images specified in .env..."
docker pull "$WEB_IMAGE"
docker pull "$ETL_IMAGE"

# Deploy stack (automatically does rolling update)
echo "🔄 Deploying stack with rolling update..."
docker stack deploy -c docker-stack.yml --with-registry-auth partijgedrag

# Monitor deployment
echo "⏳ Monitoring deployment status..."
timeout 120 bash -c '
  while true; do
    REPLICAS=$(docker service ls --filter "name=partijgedrag_web" --format "{{.Replicas}}")
    echo "   Web service: $REPLICAS"

    CURRENT=$(echo "$REPLICAS" | cut -d "/" -f 1)
    DESIRED=$(echo "$REPLICAS" | cut -d "/" -f 2)

    if [ "$CURRENT" = "$DESIRED" ] && [ "$CURRENT" -gt "0" ]; then
      echo "✅ Service has converged!"
      break
    fi

    sleep 5
  done
'

echo "✅ Deployment complete!"
echo ""
echo "📊 Service status:"
docker service ls --filter "label=com.docker.stack.namespace=partijgedrag"
