package data

import (
	"github.com/shopsilo/go-data/cache"
)

// Version of the unified go-data suite.
const Version = "1.0.0"

// Re-exported types for top-level convenience.
type CacheEngine = cache.Engine
type CacheStore = cache.Store

// NewCache initializes a cache engine.
func NewCache(store cache.Store) (*cache.Engine, error) {
	return cache.New(store)
}
