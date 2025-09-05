package renderer

import (
	"regexp"
	"sync"
)

// RegexCache provides thread-safe caching for compiled regular expressions
type RegexCache struct {
	cache map[string]*regexp.Regexp
	mu    sync.RWMutex
}

// NewRegexCache creates a new regex cache
func NewRegexCache() *RegexCache {
	return &RegexCache{
		cache: make(map[string]*regexp.Regexp),
	}
}

// Get returns a compiled regex, caching it if not already cached
func (rc *RegexCache) Get(pattern string) *regexp.Regexp {
	rc.mu.RLock()
	if regex, exists := rc.cache[pattern]; exists {
		rc.mu.RUnlock()
		return regex
	}
	rc.mu.RUnlock()

	rc.mu.Lock()
	defer rc.mu.Unlock()
	
	// Double-check after acquiring write lock
	if regex, exists := rc.cache[pattern]; exists {
		return regex
	}

	regex := regexp.MustCompile(pattern)
	rc.cache[pattern] = regex
	return regex
}

// Clear clears the cache
func (rc *RegexCache) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cache = make(map[string]*regexp.Regexp)
}
