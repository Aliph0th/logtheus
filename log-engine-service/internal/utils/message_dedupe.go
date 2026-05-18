package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

type MessageDedupeCache struct {
	ttl     time.Duration
	maxSize int

	mu    sync.Mutex
	items map[string]time.Time
	order []string
}

func NewMessageDedupeCache(ttl time.Duration, maxSize int) *MessageDedupeCache {
	return &MessageDedupeCache{
		ttl:     ttl,
		maxSize: maxSize,
		items:   make(map[string]time.Time, maxSize),
		order:   make([]string, 0, maxSize),
	}
}

func (c *MessageDedupeCache) Seen(key string) bool {
	now := time.Now().UTC()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cleanup(now)

	seenAt, exists := c.items[key]
	if !exists {
		return false
	}

	return now.Sub(seenAt) <= c.ttl
}

func (c *MessageDedupeCache) MarkSeen(key string) {
	now := time.Now().UTC()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cleanup(now)

	if _, exists := c.items[key]; !exists {
		c.order = append(c.order, key)
	}
	c.items[key] = now

	for len(c.order) > c.maxSize {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.items, oldest)
	}
}

func (c *MessageDedupeCache) cleanup(now time.Time) {
	for len(c.order) > 0 {
		oldest := c.order[0]
		timestamp, exists := c.items[oldest]
		if !exists {
			c.order = c.order[1:]
			continue
		}
		if now.Sub(timestamp) <= c.ttl {
			break
		}

		delete(c.items, oldest)
		c.order = c.order[1:]
	}
}

func HashMessageValue(value []byte) string {
	hash := sha256.Sum256(value)
	return hex.EncodeToString(hash[:])
}
