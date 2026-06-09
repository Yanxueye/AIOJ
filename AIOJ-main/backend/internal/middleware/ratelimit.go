package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/utils"
	"golang.org/x/time/rate"
)

// perUserLimiter keeps a token-bucket limiter per user id.
//
// The map is trimmed lazily: limiters whose reservoir is full and that were
// not touched in the last hour get evicted in the background janitor so the
// map does not grow unbounded with active users.
type perUserLimiter struct {
	mu      sync.Mutex
	r       rate.Limit
	burst   int
	buckets map[uint64]*userBucket
}

type userBucket struct {
	lim      *rate.Limiter
	lastSeen time.Time
}

func newPerUserLimiter(perMinute, burst int) *perUserLimiter {
	if perMinute <= 0 {
		perMinute = 12
	}
	if burst <= 0 {
		burst = 3
	}
	pl := &perUserLimiter{
		r:       rate.Limit(float64(perMinute) / 60.0),
		burst:   burst,
		buckets: make(map[uint64]*userBucket),
	}
	go pl.janitor()
	return pl
}

func (p *perUserLimiter) allow(uid uint64) bool {
	p.mu.Lock()
	b, ok := p.buckets[uid]
	if !ok {
		b = &userBucket{lim: rate.NewLimiter(p.r, p.burst)}
		p.buckets[uid] = b
	}
	b.lastSeen = time.Now()
	p.mu.Unlock()
	return b.lim.Allow()
}

func (p *perUserLimiter) janitor() {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	for range t.C {
		cutoff := time.Now().Add(-1 * time.Hour)
		p.mu.Lock()
		for uid, b := range p.buckets {
			if b.lastSeen.Before(cutoff) {
				delete(p.buckets, uid)
			}
		}
		p.mu.Unlock()
	}
}

// PerUserRateLimit rejects excessive submissions from a single authenticated
// user. Non-authenticated requests (no CurrentUserID) are allowed to pass;
// the JWT middleware is expected to run earlier in the chain.
func PerUserRateLimit(perMinute, burst int) gin.HandlerFunc {
	lim := newPerUserLimiter(perMinute, burst)
	return func(c *gin.Context) {
		uid, ok := CurrentUserID(c)
		if !ok {
			c.Next()
			return
		}
		if !lim.allow(uid) {
			utils.TooManyRequests(c, "提交过于频繁，请稍后再试")
			return
		}
		c.Next()
	}
}
