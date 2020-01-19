package cache

import (
	"TL-Gateway/config"
	"TL-Gateway/log"
	"time"
	"github.com/go-redis/redis"
)

// RedisCache implements the Cache interface
type RedisCache struct {
	redisClient *redis.Client // redis cache for token
}

// NewRedisCache return a redis cache for token
func NewRedisCache(settings *config.Config) *RedisCache {
	rc := &RedisCache{}

	// init the redis connection for token cache
	redisClient := redis.NewClient(&redis.Options{
		Addr:     settings.Redis.Addr,
		Password: settings.Redis.Passwd, // no password
		DB:       settings.Redis.DB,     // use default DB
	})
	// try to ping the redis
	if _, err := redisClient.Ping().Result(); err != nil {
		panic(err)
	}
	rc.redisClient = redisClient

	return rc
}

// Check if the key is existed and value is matched
func (rc *RedisCache) Check(k string, v string) bool {
	// check if the key is existed in redis cache
	token, err := rc.redisClient.Get(k).Result()
	if err != nil {
		return false
	}
	// check if the value is equal
	if token == v {
		return true
	}

	return false
}

// Set updates the key with value and expire time
func (rc *RedisCache) Set(k string, v string, expire int64) {
	secs := expire - time.Now().Unix() - 60
	if secs > 0 {
		// set the token to redis cache
		result := rc.redisClient.Set(k, v, time.Duration(secs)*time.Second)
		if result.Err() != nil {
			log.Errorf("update the token: %v", result.Err())
		}
	}
}
