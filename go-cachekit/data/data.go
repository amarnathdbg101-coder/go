package data

import (
	"github.com/amarnathdbg101-coder/go/go-cachekit/data/cache"
)

const Version = "1.0.0"

type CacheEngine = cache.Engine
type CacheStore = cache.Store

func NewCache(store cache.Store) (*cache.Engine, error) {
	return cache.New(store)
}
