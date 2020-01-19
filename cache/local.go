package cache

import (
	gcache "github.com/patrickmn/go-cache"
	"time"
)

const (
	// DefaultWriteTimeout - the timeout for writing the messages to queue
	DefaultWriteTimeout = 2
	// DefaultCacheExpireTime - the default expire time for local cache
	DefaultCacheExpireTime = 60 * 60
	// DefaultCacheCleanTime - the default clean time for local cache
	DefaultCacheCleanTime = 60 * 60 * 2
)

// LocalCache implements the Cache interface
type LocalCache struct {
	buffer *gcache.Cache // local cache for token
}

// NewLocalCache returns the local cache for token
func NewLocalCache() *LocalCache {
	return &LocalCache{
		buffer: gcache.New(DefaultCacheExpireTime*time.Second, DefaultCacheCleanTime*time.Second),
	}
}

// Check if the key is existed and value is matched
func (lc *LocalCache) Check(k string, v string) bool {
	// check if the key is existed in local cache
	token, ok := lc.buffer.Get(k)
	if !ok {
		return false
	}
	// check if the value is equal
	if token == v {
		return true
	}

	return false
}

// Set updates the key with value and expire time
func (lc *LocalCache) Set(k string, v string, expire int64) {
	secs := expire - time.Now().Unix() - 60
	if secs > 0 {
		// set the token to local cache
		lc.buffer.Set(k, v, time.Duration(secs)*time.Second)
	}
}
