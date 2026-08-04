package cache

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCacheSetGet(t *testing.T) {
	c := New()
	c.Set("key1", "val1")

	val, ok := c.Get("key1")
	if !ok || val != "val1" {
		t.Fatalf("expected val1, got %v (ok=%t)", val, ok)
	}

	_, ok = c.Get("key2")
	if ok {
		t.Fatalf("expected key2 to be missing")
	}
}

func TestCacheInvalidate(t *testing.T) {
	c := New()
	ver1 := c.GetSyncVersion()
	c.Set("key1", "val1")

	time.Sleep(2 * time.Millisecond)
	c.Invalidate()

	ver2 := c.GetSyncVersion()
	if ver2 <= ver1 {
		t.Fatalf("expected sync version to increase from %d, got %d", ver1, ver2)
	}

	_, ok := c.Get("key1")
	if ok {
		t.Fatalf("expected key1 to be purged after invalidation")
	}
}

func TestCacheMiddlewareETag304(t *testing.T) {
	c := New()

	handlerCallCount := 0
	handler := c.Middleware(PolicyDynamic, func(w http.ResponseWriter, r *http.Request) {
		handlerCallCount++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello world"))
	})

	// 1. Initial GET request
	req1 := httptest.NewRequest("GET", "/party-likeness?period=schoof", nil)
	rec1 := httptest.NewRecorder()
	handler(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec1.Code)
	}
	etag := rec1.Header().Get("ETag")
	if etag == "" {
		t.Fatalf("expected ETag header to be set")
	}
	cacheCtrl := rec1.Header().Get("Cache-Control")
	if cacheCtrl != "public, max-age=300, stale-while-revalidate=3600" {
		t.Fatalf("expected dynamic Cache-Control header, got %q", cacheCtrl)
	}
	if handlerCallCount != 1 {
		t.Fatalf("expected handler to be called 1 time, got %d", handlerCallCount)
	}

	// 2. Subsequent request with matching If-None-Match header -> Should return 304 Not Modified
	req2 := httptest.NewRequest("GET", "/party-likeness?period=schoof", nil)
	req2.Header.Set("If-None-Match", etag)
	rec2 := httptest.NewRecorder()
	handler(rec2, req2)

	if rec2.Code != http.StatusNotModified {
		t.Fatalf("expected status 304 Not Modified, got %d", rec2.Code)
	}
	if handlerCallCount != 1 {
		t.Fatalf("expected handler NOT to be called on 304, got count %d", handlerCallCount)
	}

	// 3. After invalidation, ETag should change and request with old ETag returns 200 OK
	c.Invalidate()
	rec3 := httptest.NewRecorder()
	handler(rec3, req2)

	if rec3.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK after invalidation, got %d", rec3.Code)
	}
	if handlerCallCount != 2 {
		t.Fatalf("expected handler to be called after invalidation, got count %d", handlerCallCount)
	}
}

func TestCacheMiddlewareImmutablePolicy(t *testing.T) {
	c := New()
	handler := c.Middleware(PolicyImmutable, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/compass/results/abc-123", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	got := rec.Header().Get("Cache-Control")
	want := "public, max-age=31536000, immutable"
	if got != want {
		t.Fatalf("expected Cache-Control %q, got %q", want, got)
	}
}
