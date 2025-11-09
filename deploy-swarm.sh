#!/bin/bash
set -e

echo "🚀 Starting Docker Swarm deployment..."

# Load environment variables
# Load environment variables
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
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
  until [ $(docker service ls --filter "name=partijgedrag_web" --format "{{.Replicas}}" | grep -o "^[0-9]*") -eq $(docker service ls --filter "name=partijgedrag_web" --format "{{.Replicas}}" | grep -o "[0-9]*$") ]; do
    echo "Waiting for services to converge..."
    docker service ls --filter "label=com.docker.stack.namespace=partijgedrag"
    sleep 5
  done
'

echo "✅ Deployment complete!"
echo ""
echo "📊 Service status:"
docker service ls --filter "label=com.docker.stack.namespace=partijgedrag"
