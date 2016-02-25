// Package inmem provides an in memory LRU cache with TTL support.
package inmem

import (
	"container/list"
	"sync"
	"time"
)

// Cache of things.
type Cache interface {
	Add(key, value interface{}, expiresAt time.Time)
	Get(key interface{}) (interface{}, bool)
	Remove(key interface{})
	Len() int
}

// cache implements a non-thread safe fixed size cache.
type cache struct {
	size  int
	lru   *list.List
	items map[interface{}]*list.Element
}

// entry in the cache.
type entry struct {
	key       interface{}
	value     interface{}
	expiresAt time.Time
}

// NewUnlocked constructs a new Cache of the given size that is not safe for
// concurrent use. If will panic if size is not a positive integer.
func NewUnlocked(size int) Cache {
	if size <= 0 {
		panic("inmem: must provide a positive size")
	}
	return &cache{
		size:  size,
		lru:   list.New(),
		items: make(map[interface{}]*list.Element),
	}
}

func (c *cache) Add(key, value interface{}, expiresAt time.Time) {
	if ent, ok := c.items[key]; ok {
		// update existing entry
		c.lru.MoveToFront(ent)
		v := ent.Value.(*entry)
		v.value = value
		v.expiresAt = expiresAt
		return
	}

	// add new entry
	c.items[key] = c.lru.PushFront(&entry{
		key:       key,
		value:     value,
		expiresAt: expiresAt,
	})

	// remove oldest
	if c.lru.Len() > c.size {
		ent := c.lru.Back()
		if ent != nil {
			c.removeElement(ent)
		}
	}
}

func (c *cache) Get(key interface{}) (interface{}, bool) {
	if ent, ok := c.items[key]; ok {
		v := ent.Value.(*entry)

		if v.expiresAt.After(time.Now()) {
			// found good entry
			c.lru.MoveToFront(ent)
			return v.value, true
		}

		// ttl expired
		c.removeElement(ent)
	}
	return nil, false
}

func (c *cache) Remove(key interface{}) {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
	}
}

func (c *cache) Len() int {
	return c.lru.Len()
}

// removeElement is used to remove a given list element from the cache
func (c *cache) removeElement(e *list.Element) {
	c.lru.Remove(e)
	kv := e.Value.(*entry)
	delete(c.items, kv.key)
}

type lockedCache struct {
	c cache
	m sync.Mutex
}

// NewLocked constructs a new Cache of the given size that is safe for
// concurrent use. If will panic if size is not a positive integer.
func NewLocked(size int) Cache {
	if size <= 0 {
		panic("inmem: must provide a positive size")
	}
	return &lockedCache{
		c: cache{
			size:  size,
			lru:   list.New(),
			items: make(map[interface{}]*list.Element),
		},
	}
}

func (l *lockedCache) Add(key, value interface{}, expiresAt time.Time) {
	l.m.Lock()
	l.c.Add(key, value, expiresAt)
	l.m.Unlock()
}

func (l *lockedCache) Get(key interface{}) (interface{}, bool) {
	l.m.Lock()
	v, f := l.c.Get(key)
	l.m.Unlock()
	return v, f
}

func (l *lockedCache) Remove(key interface{}) {
	l.m.Lock()
	l.c.Remove(key)
	l.m.Unlock()
}

func (l *lockedCache) Len() int {
	l.m.Lock()
	c := l.c.Len()
	l.m.Unlock()
	return c
}
