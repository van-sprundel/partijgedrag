#!/bin/bash
set -e

echo "🚀 Zero-Downtime Deployment (Easiest Method)"
echo ""

# This method relies on:
# 1. Running 2 app instances
# 2. Docker Compose's --wait flag for health checks
# 3. Nginx load balancing between instances

# Pull latest images
echo "📦 Pulling latest images..."
docker compose -f docker-compose.server.yml pull

# Deploy with 2 instances and wait for health checks
echo "🔄 Deploying with rolling update..."
docker compose -f docker-compose.server.yml up -d --wait --scale app=2

echo ""
echo "✅ Deployment complete!"
echo ""
echo "💡 You now have 2 app instances for redundancy."
echo "   To scale back to 1: docker compose -f docker-compose.server.yml up -d --scale app=1"
echo ""
docker compose -f docker-compose.server.yml ps
