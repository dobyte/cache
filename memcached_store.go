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
		DefaultNilExpire time.Duration
	}
)

// Create a memcached store instance.
func NewMemcachedStore(opt *MemcachedOptions) Store {
	c := &MemcachedStore{
		client: memcache.New(opt.Addrs...),
	}
	c.SetPrefix(opt.Prefix)
	c.SetDefaultNilValue(opt.DefaultNilValue)
	c.SetDefaultNilExpire(opt.DefaultNilExpire)
	
	return c
}

// Determine if an item exists in the cache.
func (c *MemcachedStore) Has(key string) (bool, error) {
	if _, err := c.client.Get(c.prefixKey(key)); err != nil {
		if err == memcache.ErrCacheMiss {
			return false, nil
		}
		
		return false, err
	}
	
	return true, nil
}

// Determine if multiple item exists in the cache.
func (c *MemcachedStore) HasMany(keys ...string) (map[string]bool, error) {
	var (
		ret        = make(map[string]bool)
		prefixKeys = make([]string, 0)
	)
	
	for _, key := range keys {
		prefixKeys = append(prefixKeys, c.prefixKey(key))
	}
	
	items, err := c.client.GetMulti(prefixKeys)
	if err != nil {
		for _, key := range keys {
			ret[key] = false
		}
	} else {
		for _, key := range keys {
			if _, ok := items[c.prefixKey(key)]; ok {
				ret[key] = true
			} else {
				ret[key] = false
			}
		}
	}
	
	return ret, err
}

// Retrieve an item from the cache by key.
func (c *MemcachedStore) Get(key string, defaultValue ...interface{}) Result {
	item, err := c.client.Get(c.prefixKey(key))
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

// Retrieve multiple items from the cache by key.
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
			if item, ok := items[c.prefixKey(key)]; ok {
				ret[key] = string(item.Value)
			} else {
				ret[key] = ""
			}
		}
	}
	
	return ret, err
}

// Retrieve or set an item from the cache by key.
func (c *MemcachedStore) GetSet(key string, fn defaultValueFunc) Result {
	if item, err := c.client.Get(c.prefixKey(key)); err != nil {
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
				return NewResult("", Nil, c.Set(key, c.GetDefaultNilValue(), c.GetDefaultNilExpire()))
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

// Store an item in the cache.
func (c *MemcachedStore) Set(key string, value interface{}, expire time.Duration) error {
	return c.client.Set(&memcache.Item{
		Key:        c.prefixKey(key),
		Value:      []byte(conv.String(value)),
		Expiration: int32(expire / time.Second),
	})
}

// Store multiple items in the cache for a given number of expire,Non-atomic operation
func (c *MemcachedStore) SetMany(values map[string]interface{}, expire time.Duration) error {
	for key, value := range values {
		if err := c.Set(key, value, expire); err != nil {
			return err
		}
	}
	
	return nil
}

// Store an item in the cache indefinitely.
func (c *MemcachedStore) Forever(key string, value interface{}) error {
	return c.Set(key, value, 0)
}

// Store multiple items in the cache indefinitely.
func (c *MemcachedStore) ForeverMany(values map[string]interface{}) error {
	return c.SetMany(values, 0)
}

// Store an item in the cache if the key does not exist.
func (c *MemcachedStore) Add(key string, value interface{}, expire time.Duration) (bool, error) {
	if err := c.client.Add(&memcache.Item{
		Key:        c.prefixKey(key),
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

// Decrement the value of an item in the cache.
func (c *MemcachedStore) Increment(key string, value int64) (int64, error) {
	if value < 0 {
		return c.Decrement(key, 0-value)
	}
	
	newValue, err := c.client.Increment(c.prefixKey(key), uint64(value))
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

// Increment the value of multiple items in the cache,Non-atomic operation
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

// Decrement the value of an item in the cache.
func (c *MemcachedStore) Decrement(key string, value int64) (int64, error) {
	if value < 0 {
		return c.Increment(key, 0-value)
	}
	
	newValue, err := c.client.Decrement(c.prefixKey(key), uint64(value))
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

// Decrement the value of multiple items in the cache,Non-atomic operation
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

// Remove an item from the cache.
func (c *MemcachedStore) Forget(key string) error {
	if err := c.client.Delete(c.prefixKey(key)); err != nil && err != memcache.ErrCacheMiss {
		return err
	}
	
	return nil
}

// Remove multiple items from the cache,Non-atomic operation
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

// Remove all items from the cache.
func (c *MemcachedStore) Flush() error {
	return c.client.DeleteAll()
}

// Get a lock instance.
func (c *MemcachedStore) Lock(name string, time time.Duration) Lock {
	return NewMemcachedLock(c.client, c.prefixKey(name), time)
}

// Get a client instance.
func (c *MemcachedStore) GetClient() interface{} {
	return c.client
}