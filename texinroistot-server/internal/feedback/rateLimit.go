package feedback

import (
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	rateLimitMax    = 5
	rateLimitWindow = 10 * time.Minute
)

// sendLog remembers when each sender last sent feedback.
type sendLog struct {
	mu    sync.Mutex
	sends map[string][]time.Time
}

// allow records a send for the key at now, unless the key already sent max
// messages within the window before now.
func (l *sendLog) allow(key string, now time.Time, max int, window time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	for k, times := range l.sends {
		recent := times[:0]
		for _, t := range times {
			if now.Sub(t) < window {
				recent = append(recent, t)
			}
		}
		if len(recent) == 0 {
			delete(l.sends, k)
		} else {
			l.sends[k] = recent
		}
	}

	if len(l.sends[key]) >= max {
		return false
	}
	l.sends[key] = append(l.sends[key], now)
	return true
}

var sent = &sendLog{sends: map[string][]time.Time{}}

// senderKey is the sender's address. The frontend proxy passes it in
// X-Forwarded-For. Fiber reuses the request's buffers, so the key is copied
// before it is kept in the send log.
func senderKey(c *fiber.Ctx) string {
	if forwarded := strings.TrimSpace(strings.Split(c.Get(fiber.HeaderXForwardedFor), ",")[0]); forwarded != "" {
		return strings.Clone(forwarded)
	}
	return strings.Clone(c.IP())
}

// RateLimit lets each sender send a few messages in a while.
func RateLimit(c *fiber.Ctx) error {
	if !sent.allow(senderKey(c), time.Now(), rateLimitMax, rateLimitWindow) {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many messages"})
	}
	return c.Next()
}
