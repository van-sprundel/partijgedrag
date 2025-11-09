# Zero-Downtime Deployment Guide

This guide explains how to deploy Partijgedrag with zero downtime on CYSO infrastructure.

## Current Setup

- **Hosting:** CYSO VPS
- **Reverse Proxy:** CYSO-managed Nginx (with load balancing)
- **Container Runtime:** Docker + Docker Compose
- **Current Deployment:** Manual SSH + `docker compose down && up` (causes downtime)

## Architecture

```
┌─────────────────────────────────────┐
│   CYSO Nginx (Load Balancer)        │
│   - Handles SSL termination          │
│   - Load balancing between instances │
│   - Health checking                  │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   Docker Compose (Your Server)      │
│                                      │
│   ┌──────────┐  ┌──────────┐        │
│   │  App #1  │  │  App #2  │        │
│   │  :8080   │  │  :8081   │        │
│   └──────────┘  └──────────┘        │
│                                      │
│   ┌──────────┐                       │
│   │   ETL    │  (cron-based)         │
│   └──────────┘                       │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   PostgreSQL Database                │
│   (Managed separately)               │
└─────────────────────────────────────┘
```

## Zero-Downtime Deployment Methods

### Method 1: Simplest (Recommended)

**File:** `deploy-easiest.sh`

```bash
./deploy-easiest.sh
```

**How it works:**
1. Pulls latest Docker images
2. Scales to 2 app instances
3. Docker Compose does rolling restart (stops old, starts new, one at a time)
4. Nginx automatically routes traffic to healthy instances

**Pros:**
- ✅ One command
- ✅ Built into Docker Compose (no extra tools)
- ✅ Automatic health checking with `--wait` flag

**Cons:**
- ⚠️ Requires Docker Compose v2.x (check with `docker compose version`)

### Method 2: Manual Control

**File:** `deploy-simple.sh`

```bash
./deploy-simple.sh
```

**How it works:**
1. Scale to 2 instances (old version still running)
2. Force recreate both (rolling restart)
3. Wait for health checks
4. Optionally scale back to 1

**Pros:**
- ✅ More control over each step
- ✅ Can verify health before proceeding
- ✅ Choose to keep 1 or 2 instances

### Method 3: Step-by-Step (Most Control)

**File:** `deploy-server.sh`

Manually orchestrates each container update with health checks between steps.

**Use when:**
- You want maximum control
- Debugging deployment issues
- Understanding the internals

### Method 4: Automated CI/CD

**File:** `.github/workflows/deploy-production.yml`

Automatically deploys when you push to `main` branch.

**Setup:**
1. Add GitHub Secrets (Settings → Secrets → Actions):
   - `CYSO_HOST` - Your server IP/hostname
   - `CYSO_USER` - SSH username
   - `CYSO_SSH_KEY` - Private SSH key (generate with `ssh-keygen`)
   - `CYSO_SSH_PORT` - SSH port (usually 22)
   - `DATABASE_URL` - PostgreSQL connection string
   - `CORS_ORIGIN` - Your frontend domain

2. Push to main:
   ```bash
   git push origin main
   ```

3. GitHub Actions will:
   - Build Docker images
   - Push to GitHub Container Registry
   - SSH into your server
   - Run zero-downtime deployment
   - Verify deployment succeeded

## Database Migrations

**CRITICAL:** Run migrations BEFORE deploying new code.

```bash
# On your server, before deployment:
cd /path/to/partijgedrag
docker compose -f docker-compose.server.yml run --rm app npm run db:migrate

# Then deploy:
./deploy-easiest.sh
```

### Migration Best Practices

**❌ Breaking changes (NEVER do this):**
```sql
-- Renaming column in same deploy
ALTER TABLE parties RENAME COLUMN name TO full_name;
-- Old code breaks immediately!
```

**✅ Backward-compatible approach:**
```sql
-- Step 1 (Deploy 1): Add new column
ALTER TABLE parties ADD COLUMN full_name TEXT;

-- Step 2 (Deploy 1): Update code to write to both columns
-- (Deploy this version)

-- Step 3 (Later): Backfill data
UPDATE parties SET full_name = name WHERE full_name IS NULL;

-- Step 4 (Deploy 2): Update code to read from full_name only
-- (Deploy this version)

-- Step 5 (Deploy 3): Remove old column
ALTER TABLE parties DROP COLUMN name;
```

## Health Checks

The app now has two health endpoints:

### `/health` - Liveness Probe
- Returns `200` if app is running
- Doesn't check dependencies
- Used by Docker for container health

### `/ready` - Readiness Probe
- Returns `200` if app is ready to serve traffic
- Checks database connection
- Returns `503` if database is down
- **Use this for zero-downtime deploys**

Test:
```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

## Graceful Shutdown

The app already handles `SIGTERM` properly:

1. Receives `SIGTERM` from Docker
2. Stops accepting new connections
3. Finishes processing existing requests (max 30s)
4. Closes database connections
5. Exits cleanly

This ensures no requests are dropped during deployment.

## Rollback Strategy

### If deployment fails during update:

**Compose method:**
```bash
# Pull previous version
docker pull ghcr.io/van-sprundel/partijgedrag-web:previous-tag

# Update compose file to use previous tag
# Then redeploy
./deploy-easiest.sh
```

**Quick rollback:**
```bash
# Stop all new containers
docker compose -f docker-compose.server.yml down

# Start previous containers (if still present)
docker compose -f docker-compose.server.yml up -d
```

### If deployment succeeds but app has bugs:

1. Revert git commit
2. Push to main (triggers new build)
3. Re-deploy

## Monitoring Deployment

### Watch logs during deployment:
```bash
# Terminal 1: Run deployment
./deploy-easiest.sh

# Terminal 2: Watch logs
docker compose -f docker-compose.server.yml logs -f app
```

### Check container health:
```bash
docker compose -f docker-compose.server.yml ps

# Should show:
# app-1   running (healthy)
# app-2   running (healthy)
```

### Test from outside:
```bash
# From your local machine
curl -I https://your-domain.nl/health
# Should return 200 OK during entire deployment
```

## Troubleshooting

### Health check fails
```bash
# Check app logs
docker compose -f docker-compose.server.yml logs app

# Test health endpoint
docker compose -f docker-compose.server.yml exec app wget -O- http://localhost:80/ready

# Check database connectivity
docker compose -f docker-compose.server.yml exec app sh -c 'echo "SELECT 1" | psql $DATABASE_URL'
```

### Deployment hangs
```bash
# Check container status
docker compose -f docker-compose.server.yml ps

# Check what's blocking
docker compose -f docker-compose.server.yml events

# Force stop and retry
docker compose -f docker-compose.server.yml down
./deploy-easiest.sh
```

### Old containers not stopping
```bash
# List all containers
docker ps -a | grep partijgedrag

# Manually stop old ones
docker stop <container-id>
docker rm <container-id>

# Clean up
docker system prune -f
```

## Performance Tuning

### Running 2 instances permanently
```bash
# Update docker-compose.server.yml
services:
  app:
    deploy:
      replicas: 2  # Add this line

# Or use scale flag
docker compose -f docker-compose.server.yml up -d --scale app=2
```

**Benefits:**
- Better redundancy
- Load distribution (if Nginx balances between them)
- Faster deployments (one instance always available)

**Cost:**
- 2x memory usage
- 2x CPU usage

### Resource limits (optional)
```yaml
services:
  app:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

## Comparison: Docker Compose vs Docker Swarm

| Feature | Docker Compose | Docker Swarm |
|---------|----------------|--------------|
| **Setup** | Already using | `docker swarm init` |
| **Zero-downtime** | Manual scaling | Built-in |
| **Rollback** | Manual | Automatic on failure |
| **Multiple nodes** | No | Yes |
| **Complexity** | Low | Medium |
| **Best for** | Single server | Multi-server |

### Should you switch to Swarm?

**No, if:**
- You only have 1 server
- Current approach works
- You want simplicity

**Yes, if:**
- You want automatic rollbacks
- You plan to add more servers
- You want built-in orchestration

## Next Steps

1. **Add health checks to docker-compose.server.yml** (already in updated version)
2. **Test deployment in non-peak hours** (verify zero-downtime)
3. **Set up automated deployments** (GitHub Actions)
4. **Add monitoring** (track deployment success/failure)
5. **Document rollback procedures** for your team

## Quick Reference

```bash
# Manual deployment (zero-downtime)
./deploy-easiest.sh

# Check logs
docker compose -f docker-compose.server.yml logs -f app

# Check health
curl http://localhost:8080/ready

# Scale to 2 instances
docker compose -f docker-compose.server.yml up -d --scale app=2

# Rollback (manual)
docker compose -f docker-compose.server.yml down
# ... restore previous version ...
docker compose -f docker-compose.server.yml up -d
```
