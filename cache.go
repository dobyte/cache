/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/5/23 2:36 下午
 * @Desc: cache interface defined
 */

package cache

import (
	"context"
	"time"
)

type Cache interface {
	// Has Determine if an item exists in the cache.
	Has(ctx context.Context, key string) (bool, error)
	// HasMany Determine if multiple item exists in the cache.
	HasMany(ctx context.Context, keys ...string) (map[string]bool, error)
	// Get Retrieve an item from the cache by key.
	Get(key string, defaultValue ...interface{}) Result
	// GetMany Retrieve multiple items from the cache by key.
	GetMany(ctx context.Context, keys ...string) (map[string]Result, error)
	// GetSet Retrieve or set an item from the cache by key.
	GetSet(key string, fn func() (interface{}, time.Duration, error)) Result
	// Set Store an item in the cache.
	Set(key string, value interface{}, expire time.Duration) error
	// SetMany Store multiple items in the cache for a given number of expire.
	SetMany(values map[string]interface{}, expire time.Duration) error
	// Forever Store an item in the cache indefinitely.
	Forever(key string, value interface{}) error
	// ForeverMany Store multiple items in the cache indefinitely.
	ForeverMany(values map[string]interface{}) error
	// Add Store an item in the cache if the key does not exist.
	Add(key string, value interface{}, expire time.Duration) (bool, error)
	// Increment Increment the value of an item in the cache.
	Increment(key string, value int64) (int64, error)
	// IncrementMany Increment the value of multiple items in the cache.
	IncrementMany(values map[string]int64) (map[string]int64, error)
	// Decrement Decrement the value of an item in the cache.
	Decrement(key string, value int64) (int64, error)
	// DecrementMany Decrement the value of multiple items in the cache.
	DecrementMany(values map[string]int64) (map[string]int64, error)
	// Forget Remove an item from the cache.
	Forget(key string) error
	// ForgetMany Remove multiple items from the cache.
	ForgetMany(keys ...string) (int64, error)
	// Expire Set expiration time for a key.
	Expire(key string, expire time.Duration) (bool, error)
	// ExpireMany Set expiration time for multiple key.
	ExpireMany(values map[string]time.Duration) (map[string]bool, error)
	// Flush Remove all items from the cache.
	Flush() error
	// Lock Get a lock instance.
	Lock(name string, time time.Duration) Lock
	// PrefixKey Add prefix to the front of key.
	PrefixKey(key string) string
	// GetClient Get a client instance.
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
		DefaultNilExpire int64
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

// Has Determine if an item exists in the cache.
func (c *cache) Has(ctx context.Context, key string) (bool, error) {
	val, err := storeSharedCallGroup.Call(key, func() (interface{}, error) {
		return c.store.Has(ctx, key)
	})

	return val.(bool), err
}

// HasMany Determine if multiple item exists in the cache.
func (c *cache) HasMany(ctx context.Context, keys ...string) (map[string]bool, error) {
	return c.store.HasMany(ctx, keys...)
}

// Get Retrieve an item from the cache by key.
func (c *cache) Get(key string, defaultValue ...interface{}) Result {
	rst, _ := storeSharedCallGroup.Call(key, func() (interface{}, error) {
		return c.store.Get(context.Background(), key, defaultValue...), nil
	})

	return rst.(Result)
}

// GetMany Retrieve multiple items from the cache by key.
func (c *cache) GetMany(ctx context.Context, keys ...string) (map[string]Result, error) {
	return c.store.GetMany(ctx, keys...)
}

// GetSet Retrieve or set an item from the cache by key.
func (c *cache) GetSet(key string, fn func() (interface{}, time.Duration, error)) Result {
	return c.store.GetSet(key, fn)
}

// Set Store an item in the cache.
func (c *cache) Set(key string, value interface{}, expire time.Duration) error {
	return c.store.Set(key, value, expire)
}

// SetMany Store multiple items in the cache for a given number of expire.
func (c *cache) SetMany(values map[string]interface{}, expire time.Duration) error {
	return c.store.SetMany(values, expire)
}

// Forever Store an item in the cache indefinitely.
func (c *cache) Forever(key string, value interface{}) error {
	return c.store.Forever(key, value)
}

// ForeverMany Store multiple items in the cache indefinitely.
func (c *cache) ForeverMany(values map[string]interface{}) error {
	return c.store.ForeverMany(values)
}

// Add Store an item in the cache if the key does not exist.
func (c *cache) Add(key string, value interface{}, expire time.Duration) (bool, error) {
	return c.store.Add(key, value, expire)
}

// Increment Increment the value of an item in the cache.
func (c *cache) Increment(key string, value int64) (int64, error) {
	return c.store.Increment(key, value)
}

// IncrementMany Increment the value of multiple items in the cache.
func (c *cache) IncrementMany(values map[string]int64) (map[string]int64, error) {
	return c.store.IncrementMany(values)
}

// Decrement Decrement the value of an item in the cache.
func (c *cache) Decrement(key string, value int64) (int64, error) {
	return c.store.Decrement(key, value)
}

// DecrementMany Decrement the value of multiple items in the cache.
func (c *cache) DecrementMany(values map[string]int64) (map[string]int64, error) {
	return c.store.DecrementMany(values)
}

// Forget Remove an item from the cache.
func (c *cache) Forget(key string) error {
	return c.store.Forget(key)
}

// ForgetMany Remove multiple items from the cache.
func (c *cache) ForgetMany(keys ...string) (int64, error) {
	return c.store.ForgetMany(keys...)
}

// Expire Set expiration time for a key.
func (c *cache) Expire(key string, expire time.Duration) (bool, error) {
	return c.store.Expire(key, expire)
}

// ExpireMany Set expiration time for multiple key.
func (c *cache) ExpireMany(values map[string]time.Duration) (map[string]bool, error) {
	return c.store.ExpireMany(values)
}

// Flush Remove all items from the cache.
func (c *cache) Flush() error {
	return c.store.Flush()
}

// Lock Get a lock instance.
func (c *cache) Lock(name string, time time.Duration) Lock {
	return c.store.Lock(name, time)
}

// PrefixKey Add prefix to the front of key.
func (c *cache) PrefixKey(key string) string {
	return c.store.PrefixKey(key)
}

// GetClient Get a client instance.
func (c *cache) GetClient() interface{} {
	return c.store.GetClient()
}
