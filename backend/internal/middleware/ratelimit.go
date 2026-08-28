package middleware

import (
	"sync"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/gin-gonic/gin"
)

type fixedWindow struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string]*bucket
}

type bucket struct {
	start time.Time
	count int
}

func newFixedWindow(limit int, window time.Duration) *fixedWindow {
	fw := &fixedWindow{limit: limit, window: window, buckets: make(map[string]*bucket)}
	go func() {
		for range time.Tick(window) {
			fw.prune(time.Now())
		}
	}()
	return fw
}

func (fw *fixedWindow) prune(now time.Time) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if len(fw.buckets) < 10_000 {
		return
	}
	for key, b := range fw.buckets {
		if now.Sub(b.start) > fw.window {
			delete(fw.buckets, key)
		}
	}
}

func (fw *fixedWindow) allow(key string, now time.Time) bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	b, ok := fw.buckets[key]
	if !ok || now.Sub(b.start) >= fw.window {
		fw.buckets[key] = &bucket{start: now, count: 1}
		return true
	}
	if b.count >= fw.limit {
		return false
	}
	b.count++
	return true
}

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	fw := newFixedWindow(limit, window)
	return func(c *gin.Context) {
		if !fw.allow(c.ClientIP(), time.Now()) {
			c.AbortWithStatusJSON(429, envelope(apperrTooManyRequests()))
			return
		}
		c.Next()
	}
}

func unauthorizedEnvelope() httpx.Response {
	err := apperrUnauthorized()
	return httpx.Response{Success: false, Data: nil, Error: err}
}

func envelope(err *httpx.AppError) httpx.Response {
	return httpx.Response{Success: false, Data: nil, Error: err}
}
