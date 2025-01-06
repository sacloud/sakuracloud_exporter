// Copyright 2019-2025 The sacloud/sakuracloud_exporter Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
