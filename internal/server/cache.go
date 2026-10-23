package server

import (
	"net/http"
	"sync"
	"time"
)

type cacheEntry struct {
	body     []byte
	status   int
	cachedAt time.Time
}

// ponytail: one map + mutex with a naive size cap (cleared when full)
// and lazy TTL expiry on read; in-flight requests are deduped by a
// per-key broadcast (hand-rolled singleflight, no deps). Upgrade
// path: size-bucketed LRU + x/sync/singleflight if this ever serves
// serious multi-tenant traffic.
type cache struct {
	mu      sync.Mutex
	m       map[string]cacheEntry
	flights map[string]*flight
	ttl     time.Duration
	maxSize int
}

func newCache(ttl time.Duration, maxSize int) *cache {
	return &cache{
		m:       make(map[string]cacheEntry),
		flights: make(map[string]*flight),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// flight is a shared in-flight computation. Waiters block on done and
// then read e/err writes happen-before close(done), so no extra
// synchronization is needed (Go memory model).
type flight struct {
	done chan struct{}
	e    cacheEntry
	err  error
}

// cacheKey includes the full URI, so /trending?type=album vs
// ?type=song and per-language variants never collide keying by
// endpoint alone is exactly what made the original's trending return
// the same payload for every request.
func cacheKey(req *http.Request) string {
	return req.Method + "?" + req.URL.RequestURI()
}

func (c *cache) get(k string) (cacheEntry, bool) {
	c.mu.Lock()
	e, ok := c.m[k]
	c.mu.Unlock()
	if !ok || time.Since(e.cachedAt) > c.ttl {
		return cacheEntry{}, false
	}
	return e, true
}

// fetch returns the cached entry, joins an already-running computation
// for the same key (shared flight), or runs compute and caches success.
// state is "hit", "shared" or "miss".
func (c *cache) fetch(k string, compute func() (cacheEntry, error)) (cacheEntry, string, error) {
	if e, ok := c.get(k); ok {
		return e, "hit", nil
	}

	c.mu.Lock()
	if f, ok := c.flights[k]; ok {
		c.mu.Unlock()
		<-f.done
		return f.e, "shared", f.err
	}
	f := &flight{done: make(chan struct{})}
	c.flights[k] = f
	c.mu.Unlock()

	f.e, f.err = compute()
	if f.err == nil {
		c.set(k, f.e)
	}

	c.mu.Lock()
	delete(c.flights, k)
	c.mu.Unlock()
	close(f.done)

	return f.e, "miss", f.err
}

func (c *cache) set(k string, e cacheEntry) {
	c.mu.Lock()
	if len(c.m) >= c.maxSize {
		c.m = make(map[string]cacheEntry)
	}
	c.m[k] = cacheEntry{body: e.body, status: e.status, cachedAt: time.Now()}
	c.mu.Unlock()
}
