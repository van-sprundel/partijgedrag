#!/bin/bash
set -e

echo "🔄 Rollback to Previous Version"
echo "================================"
echo ""

# Check if we're in Swarm mode
SWARM_ACTIVE=$(docker info 2>/dev/null | grep -q "Swarm: active" && echo "true" || echo "false")

if [ "$SWARM_ACTIVE" = "true" ]; then
  echo "📊 Docker Swarm detected"
  echo ""
  echo "Option 1: Automatic rollback (rolls back to previous version)"
  echo "   docker service update --rollback partijgedrag_web"
  echo ""
  echo "Option 2: Manual rollback to specific SHA"
  echo "   Set WEB_IMAGE in .env to desired SHA and run ./deploy-production.sh"
  echo ""

  read -p "Use automatic rollback? (y/n) " -n 1 -r
  echo ""

  if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🔄 Rolling back partijgedrag_web..."
    docker service update --rollback partijgedrag_web

    echo ""
    echo "⏳ Monitoring rollback..."
    timeout 60 bash -c '
      while true; do
        STATE=$(docker service ps partijgedrag_web --format "{{.CurrentState}}" | head -n1)
        echo "   State: $STATE"

        if echo "$STATE" | grep -q "Running"; then
          echo ""
          echo "✅ Rollback complete!"
          break
        fi

        sleep 3
      done
    ' || echo "⚠️  Rollback is taking longer than expected"

    echo ""
    echo "📊 Service status:"
    docker service ps partijgedrag_web --no-trunc
    exit 0
  else
    echo "ℹ️  Manual rollback instructions:"
    echo "   1. Edit .env and set WEB_IMAGE to previous SHA"
    echo "   2. Run: ./deploy-production.sh"
    exit 0
  fi
else
  echo "📊 Docker Compose mode detected"
  echo ""
fi

# Docker Compose rollback
echo "To rollback with Docker Compose:"
echo ""
echo "1. Find the Git SHA of the version you want to rollback to:"
echo "   git log --oneline"
echo ""
echo "2. Update .env file:"
echo "   WEB_IMAGE=ghcr.io/van-sprundel/partijgedrag-web:PREVIOUS_SHA"
echo "   ETL_IMAGE=ghcr.io/van-sprundel/partijgedrag-etl:PREVIOUS_SHA"
echo ""
echo "3. Deploy the previous version:"
echo "   ./deploy-compose.sh"
echo ""
echo "📝 Example:"
echo "   # List recent images"
echo "   docker images ghcr.io/van-sprundel/partijgedrag-web"
echo ""
echo "   # Or check GitHub Container Registry:"
echo "   # https://github.com/van-sprundel/partijgedrag/pkgs/container/partijgedrag-web"
echo ""

read -p "Enter Git SHA to rollback to (or press Enter to cancel): " TARGET_SHA

if [ -z "$TARGET_SHA" ]; then
  echo "❌ Rollback cancelled"
  exit 0
fi

echo ""
echo "🔄 Rolling back to: $TARGET_SHA"
echo ""

# Update .env file
if [ -f .env ]; then
  # Backup current .env
  cp .env .env.backup

  # Update image tags
  sed -i.bak "s|WEB_IMAGE=.*|WEB_IMAGE=ghcr.io/van-sprundel/partijgedrag-web:$TARGET_SHA|" .env
  sed -i.bak "s|ETL_IMAGE=.*|ETL_IMAGE=ghcr.io/van-sprundel/partijgedrag-etl:$TARGET_SHA|" .env
  sed -i.bak "s|GIT_SHA=.*|GIT_SHA=$TARGET_SHA|" .env

  echo "✅ Updated .env file"
  echo ""
  echo "📦 New configuration:"
  grep "IMAGE=" .env
  echo ""
else
  echo "❌ Error: .env file not found"
  exit 1
fi

read -p "Proceed with rollback deployment? (y/n) " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
  echo "🚀 Deploying rollback version..."
  ./deploy-compose.sh
else
  echo "❌ Rollback cancelled"
  echo "   .env backup saved as .env.backup"
  exit 1
fi
