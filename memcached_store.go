/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/6/1 6:39 上午
 * @Desc: a memcached store instance
 */

package cache

import (
	"time"
	
	"github.com/bradfitz/gomemcache/memcache"
	
	"github.com/dobyte/cache/internal/conv"
)

type (
	Memcached      = memcache.Client
	MemcachedStore struct {
		BaseStore
		client *Memcached
	}
	MemcachedOptions struct {
		Addrs            []string
		Prefix           string
		DefaultNilValue  string
		DefaultNilExpire int64
	}
)

// NewMemcachedStore Create a memcached store instance.
func NewMemcachedStore(opt *MemcachedOptions) Store {
	c := &MemcachedStore{
		client: memcache.New(opt.Addrs...),
	}
	c.SetPrefix(opt.Prefix)
	c.SetDefaultNilValue(opt.DefaultNilValue)
	c.SetDefaultNilExpire(opt.DefaultNilExpire)
	
	return c
}

// Has Determine if an item exists in the cache.
func (c *MemcachedStore) Has(key string) (bool, error) {
	if _, err := c.client.Get(c.PrefixKey(key)); err != nil {
		if err == memcache.ErrCacheMiss {
			return false, nil
		}
		
		return false, err
	}
	
	return true, nil
}

// HasMany Determine if multiple item exists in the cache.
func (c *MemcachedStore) HasMany(keys ...string) (map[string]bool, error) {
	var (
		ret        = make(map[string]bool)
		prefixKeys = make([]string, 0)
	)
	
	for _, key := range keys {
		prefixKeys = append(prefixKeys, c.PrefixKey(key))
	}
	
	items, err := c.client.GetMulti(prefixKeys)
	if err != nil {
		for _, key := range keys {
			ret[key] = false
		}
	} else {
		for _, key := range keys {
			if _, ok := items[c.PrefixKey(key)]; ok {
				ret[key] = true
			} else {
				ret[key] = false
			}
		}
	}
	
	return ret, err
}

// Get Retrieve an item from the cache by key.
func (c *MemcachedStore) Get(key string, defaultValue ...interface{}) Result {
	item, err := c.client.Get(c.PrefixKey(key))
	if err != nil {
		if err == memcache.ErrCacheMiss {
			if len(defaultValue) > 0 {
				return NewResult(conv.String(defaultValue[0]))
			} else {
				return NewResult("", Nil)
			}
		} else {
			return NewResult("", err)
		}
	}
	
	return NewResult(string(item.Value))
}

// GetMany Retrieve multiple items from the cache by key.
func (c *MemcachedStore) GetMany(keys ...string) (map[string]string, error) {
	var (
		ret          = make(map[string]string)
		prefixedKeys = make([]string, 0)
	)
	
	for _, key := range keys {
		prefixedKeys = append(prefixedKeys, key)
	}
	
	items, err := c.client.GetMulti(prefixedKeys)
	if err != nil {
		for _, key := range keys {
			if item, ok := items[c.PrefixKey(key)]; ok {
				ret[key] = string(item.Value)
			} else {
				ret[key] = ""
			}
		}
	}
	
	return ret, err
}

// GetSet Retrieve or set an item from the cache by key.
func (c *MemcachedStore) GetSet(key string, fn defaultValueFunc) Result {
	if item, err := c.client.Get(c.PrefixKey(key)); err != nil {
		if err != memcache.ErrCacheMiss {
			return NewResult("", err)
		} else {
			switch ret, err := storeSharedCallGroup.Call(key, func() (interface{}, error) {
				val, expire, err := fn()
				return defaultValueRet{
					val:    val,
					expire: expire,
				}, err
			}); err {
			case nil:
				ret := ret.(defaultValueRet)
				val := conv.String(ret.val)
				return NewResult(val, nil, c.Set(key, val, ret.expire))
			case Nil:
				ret := ret.(defaultValueRet)
				expire := c.GetDefaultNilExpire()
				if ret.expire > 0 {
					expire = ret.expire
				}
				return NewResult("", Nil, c.Set(key, c.GetDefaultNilValue(), expire))
			default:
				return NewResult("", err)
			}
		}
	} else {
		if val := string(item.Value); val == c.GetDefaultNilValue() {
			return NewResult("", Nil)
		} else {
			return NewResult(val)
		}
	}
}

// Set Store an item in the cache.
func (c *MemcachedStore) Set(key string, value interface{}, expire time.Duration) error {
	return c.client.Set(&memcache.Item{
		Key:        c.PrefixKey(key),
		Value:      []byte(conv.String(value)),
		Expiration: int32(expire / time.Second),
	})
}

// SetMany Store multiple items in the cache for a given number of expire,Non-atomic operation
func (c *MemcachedStore) SetMany(values map[string]interface{}, expire time.Duration) error {
	for key, value := range values {
		if err := c.Set(key, value, expire); err != nil {
			return err
		}
	}
	
	return nil
}

// Forever Store an item in the cache indefinitely.
func (c *MemcachedStore) Forever(key string, value interface{}) error {
	return c.Set(key, value, 0)
}

// ForeverMany Store multiple items in the cache indefinitely.
func (c *MemcachedStore) ForeverMany(values map[string]interface{}) error {
	return c.SetMany(values, 0)
}

// Add Store an item in the cache if the key does not exist.
func (c *MemcachedStore) Add(key string, value interface{}, expire time.Duration) (bool, error) {
	if err := c.client.Add(&memcache.Item{
		Key:        c.PrefixKey(key),
		Value:      []byte(conv.String(value)),
		Expiration: int32(expire / time.Second),
	}); err != nil {
		if err == memcache.ErrNotStored {
			return false, nil
		}
		
		return false, err
	} else {
		return true, nil
	}
}

// Increment Increment the value of an item in the cache.
func (c *MemcachedStore) Increment(key string, value int64) (int64, error) {
	if value < 0 {
		return c.Decrement(key, 0-value)
	}
	
	newValue, err := c.client.Increment(c.PrefixKey(key), uint64(value))
	if err != nil {
		if err == memcache.ErrCacheMiss {
			if _, err = c.Add(key, value, 0); err != nil {
				return 0, err
			}
			
			return value, nil
		} else {
			return 0, err
		}
	}
	
	return int64(newValue), err
}

// IncrementMany Increment the value of multiple items in the cache,Non-atomic operation
func (c *MemcachedStore) IncrementMany(values map[string]int64) (map[string]int64, error) {
	var ret = make(map[string]int64)
	
	for key, value := range values {
		if newValue, err := c.Increment(key, value); err != nil {
			return ret, err
		} else {
			ret[key] = newValue
		}
	}
	
	return ret, nil
}

// Decrement Decrement the value of an item in the cache.
func (c *MemcachedStore) Decrement(key string, value int64) (int64, error) {
	if value < 0 {
		return c.Increment(key, 0-value)
	}
	
	newValue, err := c.client.Decrement(c.PrefixKey(key), uint64(value))
	if err != nil {
		if err == memcache.ErrCacheMiss {
			if _, err = c.Add(key, value, 0); err != nil {
				return 0, err
			}
			
			return value, nil
		} else {
			return 0, err
		}
	}
	
	return int64(newValue), err
}

// DecrementMany Decrement the value of multiple items in the cache,Non-atomic operation
func (c *MemcachedStore) DecrementMany(values map[string]int64) (map[string]int64, error) {
	var ret = make(map[string]int64)
	
	for key, value := range values {
		if newValue, err := c.Decrement(key, value); err != nil {
			return ret, err
		} else {
			ret[key] = newValue
		}
	}
	
	return ret, nil
}

// Forget Remove an item from the cache.
func (c *MemcachedStore) Forget(key string) error {
	if err := c.client.Delete(c.PrefixKey(key)); err != nil && err != memcache.ErrCacheMiss {
		return err
	}
	
	return nil
}

// ForgetMany Remove multiple items from the cache,Non-atomic operation
func (c *MemcachedStore) ForgetMany(keys ...string) (int64, error) {
	var count int64 = 0
	
	for _, key := range keys {
		if err := c.Forget(key); err != nil {
			return count, err
		} else {
			count++
		}
	}
	
	return count, nil
}

// Expire Set expiration time for a key.
func (c *MemcachedStore) Expire(key string, expire time.Duration) (bool, error) {
	val, err := c.client.Get(c.PrefixKey(key))
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return false, nil
		}
		
		return false, err
	}
	
	if err = c.Set(key, val, expire); err != nil {
		return false, err
	}
	
	return true, nil
}

// ExpireMany Expire Set expiration time for multiple key.
func (c *MemcachedStore) ExpireMany(values map[string]time.Duration) (map[string]bool, error) {
	var (
		ok  bool
		err error
		ret = make(map[string]bool)
	)
	
	for key, expire := range values {
		if ok, err = c.Expire(key, expire); err != nil {
			return nil, err
		} else {
			ret[key] = ok
		}
	}
	
	return ret, err
}

// Flush Remove all items from the cache.
func (c *MemcachedStore) Flush() error {
	return c.client.DeleteAll()
}

// Lock Get a lock instance.
func (c *MemcachedStore) Lock(name string, time time.Duration) Lock {
	return NewMemcachedLock(c.client, c.PrefixKey(name), time)
}

// GetClient Get the memcached client instance.
func (c *MemcachedStore) GetClient() interface{} {
	return c.client
}
