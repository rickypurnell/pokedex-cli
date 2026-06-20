package pokecache

import (
	"sync"
	"time"
)

type casheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	Interval time.Duration
	Mu       sync.Mutex
	Entry    map[string]casheEntry
}

func NewCache(interval time.Duration) *Cache {
	c := Cache{
		Interval: interval,
		Entry:    make(map[string]casheEntry),
	}
	go c.reapLoop()
	return &c
}

func (c *Cache) Add(key string, val []byte) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	c.Entry[key] = casheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if entry, ok := c.Entry[key]; ok {
		return entry.val, true
	}
	return nil, false
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.Interval)
	for range ticker.C {
		for k, v := range c.Entry {
			if v.createdAt.Add(c.Interval).Before(time.Now()) {
				delete(c.Entry, k)
			}
		}
	}
}
