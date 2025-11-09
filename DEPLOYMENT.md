# Zero-Downtime Deployment Guide

## Overview

This project uses **immutable Git SHA tags** for deterministic, repeatable deployments with proper rollback support.

## Architecture

```
GitHub Actions (CI/CD)
  ↓
Build Docker images with Git SHA tags
  ↓
Push to GitHub Container Registry
  ↓
SSH to CYSO server
  ↓
Deploy using Docker Swarm (recommended)
  or Docker Compose (best effort)
```

## Why Docker Swarm?

Docker Compose **cannot reliably provide zero-downtime deployments**. Even with health checks, scaling, and `--wait` flags, Docker Compose may still stop old containers before new ones are fully ready.

**Docker Swarm provides:**
- ✅ True rolling updates with `order: start-first` (new container starts before old stops)
- ✅ Automatic health check integration
- ✅ Built-in rollback on failure
- ✅ No infrastructure changes needed (built into Docker)
- ✅ Works on single-node setups

**Setup:** One command: `docker swarm init`

## Deployment Methods

### Method 1: Docker Swarm (RECOMMENDED)

**File:** `deploy-production.sh`

```bash
./deploy-production.sh
```

**How it works:**
1. Checks if Swarm is initialized (offers to initialize if not)
2. Pulls images with specific Git SHA tags
3. Uses `docker stack deploy` with proper rolling update config
4. Monitors deployment until services converge
5. Verifies health checks

**Zero-downtime guarantee:** Yes (via `update_config.order: start-first`)

**Rollback:** `./rollback.sh` or `docker service update --rollback partijgedrag_web`

### Method 2: Docker Compose (BEST EFFORT)

**File:** `deploy-compose.sh`

```bash
./deploy-compose.sh
```

**How it works:**
1. Pulls images with specific Git SHA tags
2. Uses `docker compose up -d --wait` (if Compose v2.20+)
3. Waits for health checks
4. Shows deployment status

**Zero-downtime guarantee:** No (may have 1-5s gap during container recreation)

**When to use:**
- You cannot use Docker Swarm for some reason
- Brief downtime (1-5s) is acceptable
- Testing/staging environments

### Method 3: Blue-Green (MANUAL)

**File:** `deploy-bluegreen.sh`

Manual blue-green deployment with explicit container management.

**Use case:** Maximum control, requires manual Nginx reconfiguration

## Image Tagging Strategy

### Immutable Git SHA Tags

Every image is tagged with both `:latest` and `:$GIT_SHA`:

```
ghcr.io/van-sprundel/partijgedrag-web:latest
ghcr.io/van-sprundel/partijgedrag-web:abc123def456  # Git SHA
```

**Why?**
- `:latest` is **non-deterministic** (changes on every push)
- Git SHA tags are **immutable** (specific version forever)
- Enables reliable rollbacks
- Deployment history is traceable to git commits

### Environment Variables

The `.env` file specifies which version to deploy:

```bash
GIT_SHA=abc123def456
WEB_IMAGE=ghcr.io/van-sprundel/partijgedrag-web:abc123def456
ETL_IMAGE=ghcr.io/van-sprundel/partijgedrag-etl:abc123def456
DATABASE_URL=postgresql://...
CORS_ORIGIN=https://your-domain.nl
```

## Health Checks

### Endpoints

**`/health`** - Liveness probe
- Returns `200` if app is running
- Does NOT check dependencies
- Used by Docker to detect crashed processes

**`/ready`** - Readiness probe (RECOMMENDED)
- Returns `200` if app is ready to serve traffic
- **Checks database connection** (`SELECT 1`)
- Returns `503` if database is unavailable
- **Use this for deployment health checks**

### Configuration

All deployment configs use `/ready`:

```yaml
healthcheck:
  test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:80/ready"]
  interval: 10s
  timeout: 5s
  retries: 3
  start_period: 40s
```

**Why `/ready` over `/health`?**

A container can be "running" (healthy process) but not "ready" (database down). Using `/ready` prevents routing traffic to containers that cannot serve requests.

## Database Migrations

**CRITICAL:** Always run migrations BEFORE deploying new code.

### Best Practices

1. **Run migrations in a separate step:**
```bash
# On server, before deployment
docker compose -f docker-compose.server.yml run --rm app npm run db:migrate
```

2. **Make migrations backward-compatible:**

**❌ WRONG (breaking change):**
```sql
-- Rename column
ALTER TABLE parties RENAME COLUMN name TO full_name;
-- Old code breaks immediately during rolling deployment!
```

**✅ CORRECT (backward-compatible):**
```sql
-- Deploy 1: Add new column
ALTER TABLE parties ADD COLUMN full_name TEXT;

-- Deploy 1: Update code to write to BOTH columns
-- (Old code still reads 'name', new code writes both)

-- Deploy 2: Backfill data
UPDATE parties SET full_name = name WHERE full_name IS NULL;

-- Deploy 2: Update code to read from 'full_name'

-- Deploy 3: Remove old column (after confirming Deploy 2 is stable)
ALTER TABLE parties DROP COLUMN name;
```

### Expand-Contract Pattern

For schema changes during rolling deployments:

1. **Expand:** Add new schema (columns, tables) without removing old
2. **Migrate:** Deploy code that works with both old and new schema
3. **Contract:** Remove old schema after all instances are updated

## Rollback Procedures

### Automatic Rollback (Swarm only)

```bash
./rollback.sh
# or
docker service update --rollback partijgedrag_web
```

**How it works:**
- Docker Swarm keeps previous task definition
- Rolls back to exact previous state
- Handles health checks automatically

### Manual Rollback (Any method)

```bash
# 1. Find previous Git SHA
git log --oneline

# 2. Update .env
WEB_IMAGE=ghcr.io/van-sprundel/partijgedrag-web:PREVIOUS_SHA
ETL_IMAGE=ghcr.io/van-sprundel/partijgedrag-etl:PREVIOUS_SHA

# 3. Redeploy
./deploy-production.sh  # or deploy-compose.sh
```

### Rollback Database Migrations

**WARNING:** Database rollbacks are risky!

```bash
# Check migration history
npm run prisma:migrate:status

# Rollback (use with caution!)
npm run prisma:migrate:resolve --rolled-back MIGRATION_NAME
```

**Better approach:** Deploy forward-fix instead of rollback

## Automated CI/CD

### GitHub Actions Workflow

**File:** `.github/workflows/deploy-production.yml`

**Triggered on:** Push to `main` branch

**Steps:**
1. Build Docker images
2. Tag with both `:latest` and `:$GITHUB_SHA`
3. Push to GitHub Container Registry
4. SSH to CYSO server
5. Create `.env` with Git SHA tags
6. Run `deploy-production.sh`
7. Verify `/ready` endpoint
8. Report success/failure

### Required GitHub Secrets

Go to: Settings → Secrets → Actions

```
CYSO_HOST          # Server IP or hostname
CYSO_USER          # SSH username (e.g., root)
CYSO_SSH_KEY       # Private SSH key (full key including -----BEGIN-----)
CYSO_SSH_PORT      # SSH port (default: 22)
DEPLOY_PATH        # Path on server (e.g., /opt/partijgedrag)
DATABASE_URL       # Full PostgreSQL connection string
CORS_ORIGIN        # Your domain (e.g., https://partijgedrag.nl)
```

### Manual Trigger

You can manually trigger deployment from GitHub Actions UI using `workflow_dispatch`.

## Monitoring Deployment

### During Deployment

**Swarm:**
```bash
# Watch service status
watch -n 2 'docker service ls'

# View task history
docker service ps partijgedrag_web --no-trunc

# Live logs
docker service logs partijgedrag_web -f
```

**Compose:**
```bash
# Watch container status
watch -n 2 'docker compose -f docker-compose.server.yml ps'

# Live logs
docker compose -f docker-compose.server.yml logs -f app
```

### Health Monitoring

```bash
# Monitor health endpoint
watch -n 1 'curl -s http://localhost:8080/ready | jq'

# Should never return non-200 during Swarm deployment
# May have brief downtime with Compose deployment
```

### Post-Deployment Verification

```bash
# Check health
curl -I http://localhost:8080/ready

# Verify version (check logs for Git SHA)
docker service logs partijgedrag_web | grep "GIT_SHA\|version"

# Check resource usage
docker stats
```

## Troubleshooting

### Health Check Fails

```bash
# Check logs
docker service logs partijgedrag_web --tail 50

# Test health endpoint inside container
docker exec $(docker ps -q -f name=partijgedrag_web) wget -O- http://localhost:80/ready

# Check database connectivity
# The /ready endpoint already checks DB, so if it fails, DB is likely the issue
```

### Deployment Stuck

**Swarm:**
```bash
# Check what's blocking
docker service ps partijgedrag_web --no-trunc

# Force update (careful!)
docker service update --force partijgedrag_web
```

**Compose:**
```bash
# Check container status
docker compose -f docker-compose.server.yml ps

# Force recreate
docker compose -f docker-compose.server.yml up -d --force-recreate app
```

### Old Containers Not Stopping

```bash
# List all containers
docker ps -a | grep partijgedrag

# Manually clean up
docker stop $(docker ps -q -f name=partijgedrag)
docker container prune -f
```

### Image Pull Fails

```bash
# Login to GitHub Container Registry
echo "$GITHUB_TOKEN" | docker login ghcr.io -u USERNAME --password-stdin

# Manually pull
docker pull ghcr.io/van-sprundel/partijgedrag-web:SHA

# Check if image exists
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/orgs/van-sprundel/packages/container/partijgedrag-web/versions
```

## Performance Tuning

### Resource Limits (Swarm)

```yaml
deploy:
  resources:
    limits:
      cpus: '1.0'
      memory: 512M
    reservations:
      cpus: '0.5'
      memory: 256M
```

### Scaling

**Swarm:**
```bash
# Scale to 3 instances
docker service scale partijgedrag_web=3

# Or update stack file with replicas: 3
```

**Compose:**
```bash
# Scale to 2 instances
docker compose -f docker-compose.server.yml up -d --scale app=2
```

### Update Strategy Tuning

```yaml
update_config:
  parallelism: 2      # Update 2 at a time (faster)
  delay: 5s           # Less delay between updates
  failure_action: rollback
  max_failure_ratio: 0.3  # Rollback if 30% fail
```

## Comparison Table

| Feature | Docker Swarm | Docker Compose | Blue-Green Manual |
|---------|--------------|----------------|-------------------|
| **Zero-downtime** | ✅ Guaranteed | ❌ Best effort | ✅ Guaranteed |
| **Complexity** | Low | Low | High |
| **Setup time** | 1 command | None (already using) | Medium |
| **Rollback** | Automatic | Manual | Manual |
| **Health checks** | Integrated | Supported | Manual |
| **Multi-node** | Yes | No | Depends |
| **Production ready** | ✅ Yes | ⚠️ Not recommended | ✅ Yes |

## Quick Reference

```bash
# Deploy production (Swarm - RECOMMENDED)
./deploy-production.sh

# Deploy production (Compose - if Swarm unavailable)
./deploy-compose.sh

# Rollback
./rollback.sh

# View logs (Swarm)
docker service logs partijgedrag_web -f

# View logs (Compose)
docker compose -f docker-compose.server.yml logs -f app

# Check health
curl http://localhost:8080/ready

# Scale service (Swarm)
docker service scale partijgedrag_web=3

# Swarm rollback
docker service update --rollback partijgedrag_web
```

## Next Steps

1. **Enable Docker Swarm:** `docker swarm init` (one-time setup)
2. **Test deployment:** Run `./deploy-production.sh` in non-peak hours
3. **Set up GitHub Secrets:** Enable automated deployments
4. **Monitor first deployment:** Watch logs and health checks
5. **Test rollback:** Verify rollback procedure works

## Additional Resources

- [Docker Swarm tutorial](https://docs.docker.com/engine/swarm/swarm-tutorial/)
- [Docker health checks](https://docs.docker.com/engine/reference/builder/#healthcheck)
- [Prisma migrations](https://www.prisma.io/docs/concepts/components/prisma-migrate)
- [Zero-downtime deployments](https://www.martinfowler.com/bliki/BlueGreenDeployment.html)
