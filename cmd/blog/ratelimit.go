package main

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type rateLimiterClient struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	clients map[string]*rateLimiterClient
	rps     rate.Limit
	burst   int
	enabled bool
}

func newRateLimiter(rps float64, burst int, enabled bool) *rateLimiter {
	return &rateLimiter{
		clients: make(map[string]*rateLimiterClient),
		rps:     rate.Limit(rps),
		burst:   burst,
		enabled: enabled,
	}
}

// evictStale removes clients that haven't been seen in longer than maxAge,
// bounding the map's memory growth. Called directly (testable without a
// timer) and periodically via startCleanup.
func (rl *rateLimiter) evictStale(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for ip, c := range rl.clients {
		if time.Since(c.lastSeen) > maxAge {
			delete(rl.clients, ip)
		}
	}
}

func (rl *rateLimiter) startCleanup(interval, maxAge time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			rl.evictStale(maxAge)
		}
	}()
}

func (rl *rateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.enabled {
			next.ServeHTTP(w, r)
			return
		}

		ip := realIP(r)

		rl.mu.Lock()
		c, ok := rl.clients[ip]
		if !ok {
			c = &rateLimiterClient{limiter: rate.NewLimiter(rl.rps, rl.burst)}
			rl.clients[ip] = c
		}
		c.lastSeen = time.Now()
		allowed := c.limiter.Allow()
		rl.mu.Unlock()

		if !allowed {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// realIP resolves the client's actual IP for a request arriving through
// Cloudflare Tunnel. Cf-Connecting-Ip is set by Cloudflare's edge and can't
// be spoofed by the client — it's checked first. X-Real-IP/X-Forwarded-For
// are checked only as fallbacks for non-Cloudflare contexts (e.g. a local
// reverse proxy in dev); a client can set either of those to an arbitrary
// value, so they must never be trusted as the sole signal in production.
func realIP(r *http.Request) string {
	if ip := r.Header.Get("Cf-Connecting-Ip"); ip != "" {
		return ip
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		first, _, _ := strings.Cut(xff, ",")
		return strings.TrimSpace(first)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
