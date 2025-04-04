package main

import (
	"sync"
	"time"
)

type InMemoryCache struct {
	data     map[string]cacheEntry
	mu       sync.RWMutex
	ttl      time.Duration
	stopChan chan struct{}
}

type cacheEntry struct {
	value     string
	expiresAt time.Time
}

func NewInMemoryCache(ttl time.Duration) *InMemoryCache {
	c := &InMemoryCache{
		data:     make(map[string]cacheEntry),
		ttl:      ttl,
		stopChan: make(chan struct{}),
	}

	// Запускаем горутину для очистки просроченных записей
	go c.cleanupWorker()
	return c
}

func (c *InMemoryCache) cleanupWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.Cleanup()
		case <-c.stopChan:
			return
		}
	}
}

func (c *InMemoryCache) Stop() {
	close(c.stopChan)
}

func (c *InMemoryCache) GetString(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists || time.Now().After(entry.expiresAt) {
		return ""
	}
	return entry.value
}

func (c *InMemoryCache) SetString(key string, val string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheEntry{
		value:     val,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *InMemoryCache) Exist(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	return exists && time.Now().Before(entry.expiresAt)
}

func (c *InMemoryCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if now.After(entry.expiresAt) {
			delete(c.data, key)
		}
	}
}
