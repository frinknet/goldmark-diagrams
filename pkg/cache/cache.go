// Package cache holds the fragment cache used by the render pipeline: the
// Cache contract, the in-memory LRU and on-disk implementations, and the
// content-addressed key they are indexed by.
package cache

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// ModuleVersion participates in the content-addressed cache key, so bumping it
// invalidates every cached fragment after a renderer change. The public
// diagrams package re-exports it.
const ModuleVersion = "v0.7.0"

// Cache is a content-addressed store of normalized <figure> HTML fragments.
type Cache interface {
	Get(key string) (html string, ok bool)
	Set(key string, html string)
}

// Key applies the spec's content-addressed formula:
// sha256(moduleVersion ":" lang ":" theme ":" attrs ":" source).
func Key(lang, theme, attrs, source string) string {
	h := sha256.New()
	for i, part := range []string{ModuleVersion, lang, theme, attrs, source} {
		if i > 0 {
			_, _ = io.WriteString(h, ":")
		}
		_, _ = io.WriteString(h, part)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// ShortHash is the 8-hex-digit digest stamped on <figure data-hash>.
func ShortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:8]
}

type lruEntry struct {
	key, value string
}

// memoryLRU is a thread-safe in-memory LRU cache with a fixed capacity.
type memoryLRU struct {
	mu    sync.Mutex
	max   int
	order *list.List
	items map[string]*list.Element
}

// NewMemoryLRU returns a thread-safe in-memory LRU cache holding at most
// maxEntries fragments. A non-positive maxEntries falls back to 512.
func NewMemoryLRU(maxEntries int) Cache {
	if maxEntries <= 0 {
		maxEntries = 512
	}
	return &memoryLRU{
		max:   maxEntries,
		order: list.New(),
		items: make(map[string]*list.Element, maxEntries),
	}
}

// Get returns the cached fragment and whether it was present.
func (c *memoryLRU) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		return el.Value.(*lruEntry).value, true
	}
	return "", false
}

// Set stores a fragment, evicting the least recently used entry when full.
func (c *memoryLRU) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		el.Value.(*lruEntry).value = value
		return
	}
	el := c.order.PushFront(&lruEntry{key: key, value: value})
	c.items[key] = el
	for c.order.Len() > c.max {
		back := c.order.Back()
		if back == nil {
			break
		}
		c.order.Remove(back)
		delete(c.items, back.Value.(*lruEntry).key)
	}
}

// diskCache keeps one file per entry so compiled diagrams survive a restart.
// Writes go through CreateTemp + Rename, so a reader never sees a partial file.
type diskCache struct {
	dir string
}

// NewDiskCache returns a cache that stores fragments as <hash>.svg files under
// dir, writing each atomically.
func NewDiskCache(dir string) Cache {
	return &diskCache{dir: dir}
}

// Get reads the fragment for key, reporting false on any read failure.
func (c *diskCache) Get(key string) (string, bool) {
	b, err := os.ReadFile(filepath.Join(c.dir, key+".svg"))
	if err != nil {
		return "", false
	}
	return string(b), true
}

// Set writes the fragment for key atomically. Failures are swallowed: the
// cache is an optimization, never a correctness requirement.
func (c *diskCache) Set(key, value string) {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return
	}
	tmp, err := os.CreateTemp(c.dir, key+".tmp-*")
	if err != nil {
		return
	}
	name := tmp.Name()
	if _, err := tmp.WriteString(value); err != nil {
		tmp.Close()
		os.Remove(name)
		return
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return
	}
	if err := os.Rename(name, filepath.Join(c.dir, key+".svg")); err != nil {
		os.Remove(name)
	}
}
