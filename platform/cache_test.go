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
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	cleanupInterval := 10 * time.Minute
	cache := newCache(cleanupInterval)
	if cache.cleanupInterval != cleanupInterval {
		t.Errorf("expected cleanupInterval %v, got %v", cleanupInterval, cache.cleanupInterval)
	}
}

func TestCache_Set(t *testing.T) {
	cleanupInterval := 10 * time.Minute
	cache := newCache(cleanupInterval)

	item := "dummy_item"
	expiresAt := time.Now().Add(1 * time.Hour)
	err := cache.set(item, expiresAt)
	if err != nil {
		t.Error(err)
	}
	if cache.item != item {
		t.Errorf("item %v, got %v", item, cache.item)
	}
	if cache.expiresAt != expiresAt {
		t.Errorf("expiresAt %v, got %v", expiresAt, cache.expiresAt)
	}
}

func TestCache_Get_ItemNotExpired(t *testing.T) {
	cleanupInterval := 10 * time.Minute
	cache := newCache(cleanupInterval)

	item := "dummy_item"
	expiresAt := time.Now().Add(1 * time.Hour)
	err := cache.set(item, expiresAt)
	if err != nil {
		t.Error(err)
	}

	cachedItem := cache.get()
	if cachedItem != item {
		t.Errorf("cachedItem %v, got %v", item, cachedItem)
	}
}

func TestCache_Get_ItemExpired(t *testing.T) {
	cleanupInterval := 1 * time.Second
	cache := newCache(cleanupInterval)

	item := "dummy_item"
	expiresAt := time.Now().Add(1 * time.Second)
	err := cache.set(item, expiresAt)
	if err != nil {
		t.Error(err)
	}

	// キャッシュの期限切れと削除が行われるのを待つ
	time.Sleep(2 * time.Second)

	cachedItem := cache.get()
	if cachedItem != nil {
		t.Errorf("cached item not cleared, got %v", cachedItem)
	}
}
