package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CachePolicy int

const (
	PolicyDynamic   CachePolicy = iota // public, max-age=300, stale-while-revalidate=3600
	PolicyImmutable                    // public, max-age=31536000, immutable
	PolicyNoStore                      // no-store, no-cache
)

type Cache struct {
	mu           sync.RWMutex
	items        map[string]any
	syncVersion  uint64
	lastSyncedAt time.Time
	appVersion   string
}

var globalCache = New()

func Global() *Cache {
	return globalCache
}

func New() *Cache {
	return &Cache{
		items:        make(map[string]any),
		syncVersion:  1,
		lastSyncedAt: time.Now().UTC(),
		appVersion:   "1.0.0",
	}
}

// Init loads the timestamp of the last successful sync from the database if available.
func (c *Cache) Init(ctx context.Context, pool *pgxpool.Pool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if pool == nil {
		return
	}

	var lastFinished *time.Time
	err := pool.QueryRow(ctx, `
		SELECT max(finished_at)
		FROM ingestion_runs
		WHERE status = 'success' AND finished_at IS NOT NULL
	`).Scan(&lastFinished)
	if err == nil && lastFinished != nil {
		c.lastSyncedAt = lastFinished.UTC()
		c.syncVersion = uint64(lastFinished.UnixNano())
	}
}

// Invalidate purges all in-memory items and bumps the sync version & timestamp.
// Call this when sync completes or data changes.
func (c *Cache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]any)
	c.syncVersion++
	c.lastSyncedAt = time.Now().UTC()
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, found := c.items[key]
	return val, found
}

func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = value
}

func (c *Cache) GetSyncVersion() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.syncVersion
}

func (c *Cache) GetLastSyncedAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastSyncedAt
}

// GenerateETag computes an ETag string based on route, params, sync version, last synced time, and app version.
func (c *Cache) GenerateETag(route string) string {
	c.mu.RLock()
	ver := c.syncVersion
	ts := c.lastSyncedAt.Unix()
	appVer := c.appVersion
	c.mu.RUnlock()

	data := fmt.Sprintf("%s:v%d:%d:app%s", route, ver, ts, appVer)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf(`W/"%s"`, hex.EncodeToString(hash[:8]))
}

// Middleware wraps an http.HandlerFunc with Cache-Control headers and ETag / 304 Not Modified validation.
func (c *Cache) Middleware(policy CachePolicy, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if policy == PolicyNoStore {
				w.Header().Set("Cache-Control", "no-store, no-cache")
			}
			next(w, r)
			return
		}

		switch policy {
		case PolicyNoStore:
			w.Header().Set("Cache-Control", "no-store, no-cache")
			next(w, r)
			return
		case PolicyImmutable:
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		case PolicyDynamic:
			w.Header().Set("Cache-Control", "public, max-age=300, stale-while-revalidate=3600")
		}

		routeKey := r.URL.RequestURI()
		etag := c.GenerateETag(routeKey)
		lastMod := c.GetLastSyncedAt().Truncate(time.Second)

		w.Header().Set("ETag", etag)
		w.Header().Set("Last-Modified", lastMod.Format(http.TimeFormat))

		// Check If-None-Match
		ifIfNoneMatch := r.Header.Get("If-None-Match")
		if ifIfNoneMatch != "" && (ifIfNoneMatch == etag || ifIfNoneMatch == "*") {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		// Check If-Modified-Since
		ifIfModSince := r.Header.Get("If-Modified-Since")
		if ifIfNoneMatch == "" && ifIfModSince != "" {
			parsedTime, err := time.Parse(http.TimeFormat, ifIfModSince)
			if err == nil && !lastMod.After(parsedTime) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		next(w, r)
	}
}
