package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/AbiXnash/4-market/internal/response"
)

type bruteEntry struct {
	times []time.Time
	mu    sync.Mutex
}

type BruteForceProtector struct {
	entries     map[string]*bruteEntry
	maxAttempts int
	window      time.Duration
	cleanupTick time.Duration
	mu          sync.RWMutex
	stopCh      chan struct{}
}

func NewBruteForceProtector(maxAttempts int, window, cleanupTick time.Duration) *BruteForceProtector {
	b := &BruteForceProtector{
		entries:     make(map[string]*bruteEntry),
		maxAttempts: maxAttempts,
		window:      window,
		cleanupTick: cleanupTick,
		stopCh:      make(chan struct{}),
	}
	go b.cleanup()
	return b
}

func (b *BruteForceProtector) Stop() {
	close(b.stopCh)
}

func (b *BruteForceProtector) Allow(key string) bool {
	b.mu.RLock()
	entry, ok := b.entries[key]
	b.mu.RUnlock()

	if !ok {
		entry = &bruteEntry{}
		b.mu.Lock()
		b.entries[key] = entry
		b.mu.Unlock()
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-b.window)

	var recent []time.Time
	for _, t := range entry.times {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= b.maxAttempts {
		entry.times = recent
		return false
	}

	entry.times = append(recent, now)
	return true
}

func (b *BruteForceProtector) Middleware(keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			if !b.Allow(key) {
				response.TooManyRequests(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (b *BruteForceProtector) cleanup() {
	ticker := time.NewTicker(b.cleanupTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			now := time.Now()
			cutoff := now.Add(-b.window)
			for key, entry := range b.entries {
				entry.mu.Lock()
				var recent []time.Time
				for _, t := range entry.times {
					if t.After(cutoff) {
						recent = append(recent, t)
					}
				}
				if len(recent) == 0 {
					delete(b.entries, key)
				} else {
					entry.times = recent
				}
				entry.mu.Unlock()
			}
			b.mu.Unlock()
		case <-b.stopCh:
			return
		}
	}
}
