package cache

// Cache defines the common interfaces
type Cache interface {
	Check(k string, v string) bool
	Set(k string, v string, expire int64)
}
