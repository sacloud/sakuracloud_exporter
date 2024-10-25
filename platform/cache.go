package platform

import (
	"errors"
	"time"
)

type cache struct {
	cleanupInterval time.Duration
	expiresAt       time.Time
	item            any
}

func newCache(cleanupInterval time.Duration) *cache {
	c := &cache{
		cleanupInterval: cleanupInterval,
	}
	go c.startCleanup()

	return c
}

func (c *cache) set(item any, expiresAt time.Time) error {
	if item == nil {
		return errors.New("item is not set")
	}
	if expiresAt.IsZero() {
		return errors.New("expiresAt is not set")
	}

	c.item = item
	c.expiresAt = expiresAt

	return nil
}

func (c *cache) get() any {
	if time.Now().After(c.expiresAt) {
		return nil
	}

	return c.item
}

func (c *cache) clear() {
	c.item = nil
	c.expiresAt = time.Time{}
}

func (c *cache) startCleanup() {
	t := time.NewTicker(c.cleanupInterval)
	defer t.Stop()

	for {
		<-t.C
		if !c.expiresAt.IsZero() && time.Now().After(c.expiresAt) {
			c.clear()
		}
	}
}
