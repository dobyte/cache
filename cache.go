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
	GetSet(ctx context.Context, key string, fn func() (interface{}, time.Duration, error)) Result
	// Set Store an item in the cache.
	Set(ctx context.Context, key string, value interface{}, expire time.Duration) error
	// SetMany Store multiple items in the cache for a given number of expire.
	SetMany(ctx context.Context, values map[string]interface{}, expire time.Duration) error
	// Forever Store an item in the cache indefinitely.
	Forever(ctx context.Context, key string, value interface{}) error
	// ForeverMany Store multiple items in the cache indefinitely.
	ForeverMany(ctx context.Context, values map[string]interface{}) error
	// Add Store an item in the cache if the key does not exist.
	Add(ctx context.Context, key string, value interface{}, expire time.Duration) (bool, error)
	// Increment Increment the value of an item in the cache.
	Increment(ctx context.Context, key string, value int64) (int64, error)
	// IncrementMany increment the value of multiple items in the cache.
	IncrementMany(ctx context.Context, values map[string]int64) (map[string]int64, error)
	// Decrement Decrement the value of an item in the cache.
	Decrement(ctx context.Context, key string, value int64) (int64, error)
	// DecrementMany Decrement the value of multiple items in the cache.
	DecrementMany(ctx context.Context, values map[string]int64) (map[string]int64, error)
	// Forget Remove an item from the cache.
	Forget(ctx context.Context, key string) error
	// ForgetMany Remove multiple items from the cache.
	ForgetMany(ctx context.Context, keys ...string) (int64, error)
	// Expire Set expiration time for a key.
	Expire(ctx context.Context, key string, expire time.Duration) (bool, error)
	// ExpireMany Set expiration time for multiple key.
	ExpireMany(ctx context.Context, values map[string]time.Duration) (map[string]bool, error)
	// Flush Remove all items from the cache.
	Flush(ctx context.Context) error
	// Lock Get a lock instance.
	Lock(ctx context.Context, name string, time time.Duration) Lock
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
func (c *cache) GetSet(ctx context.Context, key string, fn func() (interface{}, time.Duration, error)) Result {
	return c.store.GetSet(ctx, key, fn)
}

// Set Store an item in the cache.
func (c *cache) Set(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	return c.store.Set(ctx, key, value, expire)
}

// SetMany Store multiple items in the cache for a given number of expire.
func (c *cache) SetMany(ctx context.Context, values map[string]interface{}, expire time.Duration) error {
	return c.store.SetMany(ctx, values, expire)
}

// Forever Store an item in the cache indefinitely.
func (c *cache) Forever(ctx context.Context, key string, value interface{}) error {
	return c.store.Forever(ctx, key, value)
}

// ForeverMany Store multiple items in the cache indefinitely.
func (c *cache) ForeverMany(ctx context.Context, values map[string]interface{}) error {
	return c.store.ForeverMany(ctx, values)
}

// Add Store an item in the cache if the key does not exist.
func (c *cache) Add(ctx context.Context, key string, value interface{}, expire time.Duration) (bool, error) {
	return c.store.Add(ctx, key, value, expire)
}

// Increment Increment the value of an item in the cache.
func (c *cache) Increment(ctx context.Context, key string, value int64) (int64, error) {
	return c.store.Increment(ctx, key, value)
}

// IncrementMany Increment the value of multiple items in the cache.
func (c *cache) IncrementMany(ctx context.Context, values map[string]int64) (map[string]int64, error) {
	return c.store.IncrementMany(ctx, values)
}

// Decrement Decrement the value of an item in the cache.
func (c *cache) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return c.store.Decrement(ctx, key, value)
}

// DecrementMany Decrement the value of multiple items in the cache.
func (c *cache) DecrementMany(ctx context.Context, values map[string]int64) (map[string]int64, error) {
	return c.store.DecrementMany(ctx, values)
}

// Forget Remove an item from the cache.
func (c *cache) Forget(ctx context.Context, key string) error {
	return c.store.Forget(ctx, key)
}

// ForgetMany Remove multiple items from the cache.
func (c *cache) ForgetMany(ctx context.Context, keys ...string) (int64, error) {
	return c.store.ForgetMany(ctx, keys...)
}

// Expire Set expiration time for a key.
func (c *cache) Expire(ctx context.Context, key string, expire time.Duration) (bool, error) {
	return c.store.Expire(ctx, key, expire)
}

// ExpireMany Set expiration time for multiple key.
func (c *cache) ExpireMany(ctx context.Context, values map[string]time.Duration) (map[string]bool, error) {
	return c.store.ExpireMany(ctx, values)
}

// Flush Remove all items from the cache.
func (c *cache) Flush(ctx context.Context) error {
	return c.store.Flush(ctx)
}

// Lock Get a lock instance.
func (c *cache) Lock(ctx context.Context, name string, time time.Duration) Lock {
	return c.store.Lock(ctx, name, time)
}

// PrefixKey Add prefix to the front of key.
func (c *cache) PrefixKey(key string) string {
	return c.store.PrefixKey(key)
}

// GetClient Get a client instance.
func (c *cache) GetClient() interface{} {
	return c.store.GetClient()
}
