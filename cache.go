/**
 * @Author: wanglin
 * @Email: wanglin@vspn.com
 * @Date: 2021/5/23 2:36 下午
 * @Desc: cache interface defined
 */

package cache

import (
	"time"
)

type Cache interface {
	// Determine if an item exists in the cache.
	Has(key string) (bool, error)
	// Determine if multiple item exists in the cache.
	HasMany(keys ...string) (map[string]bool, error)
	// Retrieve an item from the cache by key.
	Get(key string, defaultValue ...interface{}) Result
	// Retrieve multiple items from the cache by key.
	GetMany(keys ...string) (map[string]string, error)
	// Retrieve or set an item from the cache by key.
	GetSet(key string, fn func() (interface{}, time.Duration, error)) Result
	// Store an item in the cache.
	Set(key string, value interface{}, expire time.Duration) error
	// Store multiple items in the cache for a given number of expire.
	SetMany(values map[string]interface{}, expire time.Duration) error
	// Store an item in the cache indefinitely.
	Forever(key string, value interface{}) error
	// Store multiple items in the cache indefinitely.
	ForeverMany(values map[string]interface{}) error
	// Store an item in the cache if the key does not exist.
	Add(key string, value interface{}, expire time.Duration) (bool, error)
	// Increment the value of an item in the cache.
	Increment(key string, value int64) (int64, error)
	// Increment the value of multiple items in the cache.
	IncrementMany(values map[string]int64) (map[string]int64, error)
	// Decrement the value of an item in the cache.
	Decrement(key string, value int64) (int64, error)
	// Decrement the value of multiple items in the cache.
	DecrementMany(values map[string]int64) (map[string]int64, error)
	// Remove an item from the cache.
	Forget(key string) error
	// Remove multiple items from the cache.
	ForgetMany(keys ...string) (int64, error)
	// Remove all items from the cache.
	Flush() error
	// Get a lock instance.
	Lock(name string, time time.Duration) Lock
	// Get a client instance.
	GetClient() interface{}
}

const (
	RedisDriver     = "redis"
	MemcachedDriver = "memcached"
)

type (
	Stores struct {
		Redis     *RedisOptions
		Memcached *MemcachedOptions
	}
	
	Options struct {
		Driver           string
		Prefix           string
		DefaultNilValue  string
		DefaultNilExpire time.Duration
		Stores           Stores
	}
	
	cache struct {
		store Store
	}
)

func NewCache(opt *Options) Cache {
	var store Store
	
	switch opt.Driver {
	case RedisDriver:
		store = newRedisStore(opt)
	case MemcachedDriver:
		store = newMemcachedStore(opt)
	}
	
	return &cache{
		store: store,
	}
}

// Create a redis store instance.
func newRedisStore(opt *Options) Store {
	option := &RedisOptions{
		Addrs:            opt.Stores.Redis.Addrs,
		Username:         opt.Stores.Redis.Username,
		Password:         opt.Stores.Redis.Password,
		DB:               opt.Stores.Redis.DB,
		Prefix:           opt.Prefix,
		DefaultNilValue:  opt.DefaultNilValue,
		DefaultNilExpire: opt.DefaultNilExpire,
	}
	
	if opt.Stores.Redis.Prefix != "" {
		option.Prefix = opt.Stores.Redis.Prefix
	}
	
	if opt.Stores.Redis.DefaultNilValue != "" {
		option.DefaultNilValue = opt.Stores.Redis.DefaultNilValue
	}
	
	if opt.Stores.Redis.DefaultNilExpire != 0 {
		option.DefaultNilExpire = opt.Stores.Redis.DefaultNilExpire
	}
	
	return NewRedisStore(option)
}

// Create a memcached store instance.
func newMemcachedStore(opt *Options) Store {
	option := &MemcachedOptions{
		Addrs:            opt.Stores.Memcached.Addrs,
		Prefix:           opt.Prefix,
		DefaultNilValue:  opt.DefaultNilValue,
		DefaultNilExpire: opt.DefaultNilExpire,
	}
	
	if opt.Stores.Memcached.Prefix != "" {
		option.Prefix = opt.Stores.Memcached.Prefix
	}
	
	if opt.Stores.Memcached.DefaultNilValue != "" {
		option.DefaultNilValue = opt.Stores.Memcached.DefaultNilValue
	}
	
	if opt.Stores.Memcached.DefaultNilExpire != 0 {
		option.DefaultNilExpire = opt.Stores.Memcached.DefaultNilExpire
	}
	
	return NewMemcachedStore(option)
}

// Determine if an item exists in the cache.
func (c *cache) Has(key string) (bool, error) {
	return c.store.Has(key)
}

// Determine if multiple item exists in the cache.
func (c *cache) HasMany(keys ...string) (map[string]bool, error) {
	return c.store.HasMany(keys...)
}

// Retrieve an item from the cache by key.
func (c *cache) Get(key string, defaultValue ...interface{}) Result {
	return c.store.Get(key, defaultValue...)
}

// Retrieve multiple items from the cache by key.
func (c *cache) GetMany(keys ...string) (map[string]string, error) {
	return c.store.GetMany(keys...)
}

// Retrieve or set an item from the cache by key.
func (c *cache) GetSet(key string, fn func() (interface{}, time.Duration, error)) Result {
	return c.store.GetSet(key, fn)
}

// Store an item in the cache.
func (c *cache) Set(key string, value interface{}, expire time.Duration) error {
	return c.store.Set(key, value, expire)
}

// Store multiple items in the cache for a given number of expire.
func (c *cache) SetMany(values map[string]interface{}, expire time.Duration) error {
	return c.store.SetMany(values, expire)
}

// Store an item in the cache indefinitely.
func (c *cache) Forever(key string, value interface{}) error {
	return c.store.Forever(key, value)
}

// Store multiple items in the cache indefinitely.
func (c *cache) ForeverMany(values map[string]interface{}) error {
	return c.store.ForeverMany(values)
}

// Store an item in the cache if the key does not exist.
func (c *cache) Add(key string, value interface{}, expire time.Duration) (bool, error) {
	return c.store.Add(key, value, expire)
}

// Increment the value of an item in the cache.
func (c *cache) Increment(key string, value int64) (int64, error) {
	return c.store.Increment(key, value)
}

// Increment the value of multiple items in the cache.
func (c *cache) IncrementMany(values map[string]int64) (map[string]int64, error) {
	return c.store.IncrementMany(values)
}

// Decrement the value of an item in the cache.
func (c *cache) Decrement(key string, value int64) (int64, error) {
	return c.store.Decrement(key, value)
}

// Decrement the value of multiple items in the cache.
func (c *cache) DecrementMany(values map[string]int64) (map[string]int64, error) {
	return c.store.DecrementMany(values)
}

// Remove an item from the cache.
func (c *cache) Forget(key string) error {
	return c.store.Forget(key)
}

// Remove multiple items from the cache.
func (c *cache) ForgetMany(keys ...string) (int64, error) {
	return c.store.ForgetMany(keys...)
}

// Remove all items from the cache.
func (c *cache) Flush() error {
	return c.store.Flush()
}

// Get a lock instance.
func (c *cache) Lock(name string, time time.Duration) Lock {
	return c.store.Lock(name, time)
}

// Get a client instance.
func (c *cache) GetClient() interface{} {
	return c.store.GetClient()
}
