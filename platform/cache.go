package platform

import (
	"errors"
	"sync"
	"time"
)

type cache struct {
	cleanupInterval time.Duration
	expiresAt       time.Time
	item            any
	mu              sync.Mutex
}

func newCache(cleanupInterval time.Duration) *cache {
	c := &cache{
		cleanupInterval: cleanupInterval,
	}
	go c.cleanup()

	return c
}

func (c *cache) set(item any, expiresAt time.Time) error {
	if item == nil {
		return errors.New("item is not set")
	}
	if expiresAt.IsZero() {
		return errors.New("expiresAt is not set")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.item = item
	c.expiresAt = expiresAt

	return nil
}

func (c *cache) get() any {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().After(c.expiresAt) {
		return nil
	}

	return c.item
}

func (c *cache) cleanup() {
	t := time.NewTicker(c.cleanupInterval)
	defer t.Stop()

	for {
		<-t.C
		c.mu.Lock()
		if !c.expiresAt.IsZero() && time.Now().After(c.expiresAt) {
			c.item = nil
			c.expiresAt = time.Time{}
		}
		c.mu.Unlock()
	}
}
