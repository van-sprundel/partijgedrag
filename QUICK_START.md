# Zero-Downtime Deployment - Quick Start

## The Problem

Your current deployment:
```bash
docker compose down && docker compose up
```

**Causes 5-30 seconds of downtime** because all containers stop before new ones start.

## The Solution

Use **Docker Swarm** (built into Docker) for true zero-downtime deployments.

## Quick Setup

### 1. Enable Docker Swarm (One-Time)

SSH to your CYSO server:

```bash
docker swarm init
```

That's it! Swarm is now enabled.

### 2. Deploy

```bash
cd /opt/partijgedrag  # or your deployment directory
./deploy-production.sh
```

**What happens:**
- Pulls images with immutable Git SHA tags
- Starts new containers BEFORE stopping old ones (`order: start-first`)
- Waits for health checks (`/ready` endpoint checks DB connection)
- Removes old containers
- **Zero downtime! 🎉**

## Why Docker Swarm?

**Docker Compose cannot reliably do zero-downtime deployments.** Even with health checks and `--wait`, it may stop old containers before new ones are ready.

**Docker Swarm guarantees zero-downtime:**
- ✅ Starts new container before stopping old (`order: start-first`)
- ✅ Integrated health checks
- ✅ Automatic rollback on failure
- ✅ No infrastructure changes (already in Docker)
- ✅ Works on single-node setups

## Key Improvements

### 1. Immutable Git SHA Tags

**Before:**
```bash
image: ghcr.io/van-sprundel/partijgedrag-web:latest  # ❌ Non-deterministic
```

**After:**
```bash
image: ghcr.io/van-sprundel/partijgedrag-web:abc123def  # ✅ Immutable, traceable
```

**Why?** `:latest` changes on every push. Git SHA tags are permanent and enable reliable rollbacks.

### 2. Enhanced Health Checks

**Before:**
```typescript
app.get("/health", (_req, res) => {
  res.json({ status: "ok" });  // ❌ Doesn't check dependencies
});
```

**After:**
```typescript
app.get("/ready", async (_req, res) => {
  await db.$queryRaw`SELECT 1`;  // ✅ Checks database connection
  res.json({ status: "ready", database: "connected" });
});
```

**Why?** A container can be "running" but not "ready" (e.g., DB connection failed). `/ready` prevents routing traffic to broken containers.

### 3. Proper Rolling Updates

**Docker Compose (before):**
```yaml
# Recreates containers, may stop old before new is ready
# NO guarantee of zero-downtime
```

**Docker Swarm (after):**
```yaml
update_config:
  order: start-first       # ✅ New starts before old stops
  parallelism: 1           # ✅ One at a time
  failure_action: rollback # ✅ Auto-rollback on failure
```

## Automated CI/CD

### Setup GitHub Secrets

Go to: Settings → Secrets → Actions

Add:
- `CYSO_HOST` - Server IP/hostname
- `CYSO_USER` - SSH username
- `CYSO_SSH_KEY` - Your private SSH key
- `CYSO_SSH_PORT` - SSH port (usually 22)
- `DATABASE_URL` - PostgreSQL connection string
- `CORS_ORIGIN` - Your domain (e.g., `https://partijgedrag.nl`)
- `DEPLOY_PATH` - Deployment directory (e.g., `/opt/partijgedrag`)

### Push to Main = Auto-Deploy

```bash
git add .
git commit -m "feat: add new feature"
git push origin main
```

**GitHub Actions will:**
1. Build images tagged with Git SHA
2. Push to GitHub Container Registry
3. SSH to your server
4. Deploy with zero-downtime
5. Verify `/ready` endpoint
6. Report success/failure

## Rollback

### Automatic (Swarm)

```bash
./rollback.sh
# or
docker service update --rollback partijgedrag_web
```

### Manual (Any method)

```bash
# 1. Find previous version
git log --oneline

# 2. Update .env
WEB_IMAGE=ghcr.io/van-sprundel/partijgedrag-web:previous-sha
ETL_IMAGE=ghcr.io/van-sprundel/partijgedrag-etl:previous-sha

# 3. Redeploy
./deploy-production.sh
```

## Alternative: Docker Compose (Best Effort)

If you can't use Swarm:

```bash
./deploy-compose.sh
```

**⚠️ Warning:** May have 1-5s downtime during container recreation. Use for:
- Testing/staging environments
- When brief downtime is acceptable
- Quick local testing

## Deployment Methods Comparison

| Method | Zero-Downtime | Complexity | Rollback | Recommended |
|--------|---------------|------------|----------|-------------|
| **Docker Swarm** | ✅ Guaranteed | Low (1 command setup) | Automatic | ✅ **YES** |
| **Docker Compose** | ❌ Best effort | Low (already using) | Manual | ⚠️ Staging only |
| **Blue-Green Manual** | ✅ Guaranteed | High (manual steps) | Manual | Advanced use cases |
| **Current (down + up)** | ❌ 5-30s downtime | Low | N/A | ❌ **NEVER** |

## Monitor Deployment

### During Deployment

```bash
# Terminal 1: Deploy
./deploy-production.sh

# Terminal 2: Monitor health (should NEVER fail)
watch -n 1 'curl -s http://localhost:8080/ready | jq'

# Terminal 3: Watch services
watch -n 2 'docker service ls'
```

### Post-Deployment

```bash
# Check health
curl http://localhost:8080/ready

# View logs
docker service logs partijgedrag_web -f

# Check version
docker service inspect partijgedrag_web | grep Image
```

## Database Migrations

**CRITICAL:** Run migrations BEFORE deploying!

```bash
# On server
docker compose -f docker-compose.server.yml run --rm app npm run db:migrate

# Then deploy
./deploy-production.sh
```

**Important:** Make migrations backward-compatible:
- ✅ Add new columns (keep old ones)
- ✅ Gradually migrate data
- ✅ Remove old columns in next release
- ❌ Don't rename columns in same deploy as code change

See `DEPLOYMENT.md` for detailed migration strategies.

## FAQ

**Q: Do I need multiple servers for Swarm?**
A: No! Swarm works perfectly on single-node setups. It's just an orchestration mode, not a clustering requirement.

**Q: Will Swarm change my existing setup?**
A: Minimal. You just run `docker swarm init` once, then use `docker stack deploy` instead of `docker compose up`.

**Q: What about my existing Docker Compose setup?**
A: It keeps working! Docker Compose and Docker Swarm coexist. You can use both.

**Q: Can I rollback database migrations?**
A: Database rollbacks are risky. Better approach: deploy a forward-fix. See `DEPLOYMENT.md`.

**Q: How do I scale to more instances?**
A:
```bash
docker service scale partijgedrag_web=3
```

**Q: What if deployment fails?**
A: Swarm automatically rolls back to previous version. Check logs with:
```bash
docker service logs partijgedrag_web --tail 50
```

## Summary

**Before:**
```bash
ssh server
docker compose down && docker compose up  # ❌ 30s downtime
```

**After:**
```bash
git push origin main  # ✅ Auto-deploy with 0s downtime
```

Or manually:
```bash
ssh server
./deploy-production.sh  # ✅ Zero-downtime deployment
```

## Next Steps

1. ✅ **Enable Swarm:** `docker swarm init`
2. ✅ **Test deployment:** `./deploy-production.sh`
3. ✅ **Set up GitHub Secrets** for automated deploys
4. ✅ **Test rollback:** Verify rollback procedure works
5. ✅ **Monitor first production deploy** during off-peak hours

## Read More

- **DEPLOYMENT.md** - Comprehensive deployment guide
- **deploy-production.sh** - Swarm deployment script
- **deploy-compose.sh** - Fallback Docker Compose script
- **rollback.sh** - Rollback script

---

**The Bottom Line:**

Docker Compose + `down && up` = ❌ Downtime
Docker Swarm + `stack deploy` = ✅ Zero-downtime

One command to enable: `docker swarm init`
You're welcome! 🚀
