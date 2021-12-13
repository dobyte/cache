/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/5/22 3:11 下午
 * @Desc: store interface define
 */

package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/dobyte/cache/internal/sync"
)

const (
	defaultNilValue  = "cache@nil"
	defaultNilExpire = 10 * time.Second
)

var storeSharedCallGroup = sync.NewSharedCallGroup()

type (
	defaultValueFunc = func() (interface{}, time.Duration, error)
	defaultValueRet  = struct {
		val    interface{}
		expire time.Duration
	}
)

type Store interface {
	// Has Determine if an item exists in the cache.
	Has(ctx context.Context, key string) (bool, error)
	// HasMany Determine if multiple item exists in the cache.
	HasMany(ctx context.Context, keys ...string) (map[string]bool, error)
	// Get Retrieve an item from the cache by key.
	Get(ctx context.Context, key string, defaultValue ...interface{}) Result
	// GetMany Retrieve multiple items from the cache by key.
	GetMany(ctx context.Context, keys ...string) (map[string]Result, error)
	// GetSet Retrieve or set an item from the cache by key.
	GetSet(ctx context.Context, key string, fn defaultValueFunc) Result
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
	// IncrementMany Increment the value of multiple items in the cache.
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
	Lock(name string, time time.Duration) Lock
	// PrefixKey Add prefix to the front of key.
	PrefixKey(key string) string
	// GetClient Get a client instance.
	GetClient() interface{}
}

type BaseStore struct {
	prefix           string
	defaultNilValue  string
	defaultNilExpire time.Duration
}

// GetPrefix Get the cache key prefix.
func (s *BaseStore) GetPrefix() string {
	return s.prefix
}

// SetPrefix Set the cache key prefix.
func (s *BaseStore) SetPrefix(prefix string) {
	s.prefix = prefix
}

// GetDefaultNilValue Get the cache default empty value.
func (s *BaseStore) GetDefaultNilValue() string {
	return s.defaultNilValue
}

// SetDefaultNilValue Set the cache default empty value.
func (s *BaseStore) SetDefaultNilValue(value string) {
	if value == "" {
		s.defaultNilValue = defaultNilValue
	} else {
		s.defaultNilValue = value
	}
}

// GetDefaultNilExpire Get the cache default empty value expire.
func (s *BaseStore) GetDefaultNilExpire() time.Duration {
	return s.defaultNilExpire
}

// SetDefaultNilExpire Set the cache default empty value expire.
func (s *BaseStore) SetDefaultNilExpire(expire int64) {
	if expire <= 0 {
		s.defaultNilExpire = defaultNilExpire
	} else {
		s.defaultNilExpire = time.Duration(expire) * time.Second
	}
}

// PrefixKey Add prefix to the front of key.
func (s *BaseStore) PrefixKey(key string) string {
	if s.prefix == "" {
		return key
	} else {
		return fmt.Sprintf("%s:%s", s.prefix, key)
	}
}
