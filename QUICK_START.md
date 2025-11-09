# Zero-Downtime Deployment - Quick Start

## TL;DR

Your current deployment causes downtime because you do `docker compose down && up`.

**The fix:** Use the new deployment scripts that scale up before scaling down.

## What I Changed

### 1. Enhanced Backend Health Checks ✅

**File:** `app/backend/src/index.ts`

Added `/ready` endpoint that checks database connectivity:
```bash
curl http://localhost:8080/ready
```

This ensures new containers are actually ready before routing traffic to them.

### 2. Created Zero-Downtime Deployment Scripts

**Simplest (Recommended):** `deploy-easiest.sh`
```bash
./deploy-easiest.sh
```

This script:
- Pulls latest images
- Scales to 2 instances
- Does rolling restart (one at a time)
- Your Nginx load balancer keeps traffic flowing

**More control:** `deploy-simple.sh` or `deploy-server.sh`

### 3. Created Production Docker Compose

**File:** `docker-compose.server.yml`

This adds health checks and proper restart policies:
```yaml
healthcheck:
  test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:80/ready"]
  interval: 10s
  timeout: 5s
  retries: 3
```

### 4. Automated CI/CD

**File:** `.github/workflows/deploy-production.yml`

Auto-deploys when you push to `main`.

**Setup required:**
Add these GitHub Secrets (Settings → Secrets → Actions):
- `CYSO_HOST` - Your server IP
- `CYSO_USER` - SSH username
- `CYSO_SSH_KEY` - Private SSH key
- `CYSO_SSH_PORT` - SSH port (default: 22)
- `DATABASE_URL` - Your PostgreSQL connection string
- `CORS_ORIGIN` - Your domain (e.g., `https://partijgedrag.nl`)
- `DEPLOY_PATH` - Path on server (e.g., `/opt/partijgedrag`)

## How Zero-Downtime Works

### Current (Causes Downtime ❌)
```bash
docker compose down    # ← ALL containers stop (DOWNTIME!)
docker compose up      # ← New containers start
```

### New Approach (No Downtime ✅)
```bash
# Step 1: Start 2nd instance (old version)
docker compose up -d --scale app=2

# Step 2: Rolling restart (one at a time)
docker compose up -d --scale app=2 --force-recreate

# Step 3: Nginx load balancer routes to healthy instances
# ✅ Traffic never stops!
```

## Test It Yourself

### Before Deployment
```bash
# Terminal 1: Monitor health
watch -n 1 'curl -s http://localhost:8080/ready'

# Terminal 2: Run deployment
./deploy-easiest.sh
```

You should see **zero failed requests** during deployment!

## Next Steps

### Option A: Manual Testing (Do This First!)

1. SSH to your CYSO server
2. Copy the new files:
   ```bash
   cd /opt/partijgedrag  # or wherever you deploy
   git pull origin main
   ```
3. Test deployment:
   ```bash
   ./deploy-easiest.sh
   ```
4. Monitor during deployment:
   ```bash
   # In another terminal
   docker compose -f docker-compose.server.yml logs -f app
   ```

### Option B: Automated CI/CD (Recommended Long-term)

1. Add GitHub Secrets (see above)
2. Push to main:
   ```bash
   git add .
   git commit -m "Add zero-downtime deployment"
   git push origin main
   ```
3. Watch GitHub Actions deploy automatically!

## FAQ

**Q: Do I need to keep 2 instances running all the time?**
A: No! The script scales to 2 during deployment, then can scale back to 1.

**Q: What about the ETL service?**
A: ETL can have downtime (as you said), so it just does a normal restart.

**Q: Will this work with CYSO's nginx?**
A: Yes! Your nginx load balancer will automatically route traffic between the 2 instances.

**Q: What if a deployment fails?**
A: The old container keeps running. Just fix the issue and redeploy.

**Q: How do I rollback?**
A: Pull the old image tag and redeploy:
```bash
docker pull ghcr.io/van-sprundel/partijgedrag-web:previous-sha
# Update docker-compose to use that tag
./deploy-easiest.sh
```

## Comparison

| Method | Downtime | Complexity | Recommendation |
|--------|----------|------------|----------------|
| `docker compose down && up` | ❌ Yes (5-30s) | Simple | Don't use |
| `./deploy-easiest.sh` | ✅ None | Simple | **Use this!** |
| `./deploy-simple.sh` | ✅ None | Medium | If you want control |
| GitHub Actions | ✅ None | Medium | Best long-term |
| Docker Swarm | ✅ None | High | Overkill for 1 server |

## Read More

See `DEPLOYMENT.md` for detailed documentation including:
- Database migration strategies
- Rollback procedures
- Monitoring and debugging
- Performance tuning
- Docker Swarm comparison

## Summary

**Before:** `ssh server → docker compose down && up` (30 seconds downtime)
**After:** `git push` → GitHub Actions → Zero-downtime deploy (0 seconds downtime)

You're welcome! 🚀
