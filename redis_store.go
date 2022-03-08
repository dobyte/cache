/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/5/22 3:21 下午
 * @Desc: a redis store instance
 */

package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/dobyte/cache/internal/conv"
)

type (
	Redis      = redis.UniversalClient
	RedisStore struct {
		BaseStore
		client Redis
	}

	RedisOptions struct {
		Addrs            []string
		Username         string
		Password         string
		DB               int
		Prefix           string
		DefaultNilValue  string
		DefaultNilExpire int64
	}
)

// NewRedisStore Create a redis store instance.
func NewRedisStore(opt *RedisOptions) Store {
	c := &RedisStore{client: redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    opt.Addrs,
		Username: opt.Username,
		Password: opt.Password,
		DB:       opt.DB,
	})}
	c.SetPrefix(opt.Prefix)
	c.SetDefaultNilValue(opt.DefaultNilValue)
	c.SetDefaultNilExpire(opt.DefaultNilExpire)

	return c
}

// Has Determine if an items exists in the cache.
func (c *RedisStore) Has(ctx context.Context, key string) (bool, error) {
	ret := c.client.Exists(ctx, c.PrefixKey(key))

	return ret.Val() != 0, ret.Err()
}

// HasMany Determine if multiple item exists in the cache.
func (c *RedisStore) HasMany(ctx context.Context, keys ...string) (map[string]bool, error) {
	switch len(keys) {
	case 0:
		return nil, nil
	case 1:
		if v, err := c.Has(ctx, keys[0]); err != nil {
			return nil, err
		} else {
			return map[string]bool{keys[0]: v}, nil
		}
	}

	var (
		ret          = make(map[string]bool)
		lua          = `local v = {} for _,k in ipairs(KEYS) do table.insert(v, redis.call('exists',k)) end return v`
		prefixedKeys = make([]string, len(keys))
	)

	for i, key := range keys {
		prefixedKeys[i] = c.PrefixKey(key)
	}

	rst, err := c.client.Eval(ctx, lua, prefixedKeys).Result()
	if err != nil {
		return nil, err
	}

	for i, val := range rst.([]interface{}) {
		switch v := val.(type) {
		case int64:
			ret[keys[i]] = v == 1
		}
	}

	return ret, nil
}

// Get Retrieve an item from the cache by key.
func (c *RedisStore) Get(ctx context.Context, key string, defaultValue ...interface{}) Result {
	val, err := c.client.Get(ctx, c.PrefixKey(key)).Result()
	if err == redis.Nil {
		if len(defaultValue) > 0 {
			return NewResult(conv.String(defaultValue[0]))
		}

		return NewResult("", Nil)
	}

	return NewResult(val, err)
}

// GetMany Retrieve multiple items from the cache by key.
func (c *RedisStore) GetMany(ctx context.Context, keys ...string) (map[string]Result, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	var (
		ret          = make(map[string]Result, len(keys))
		prefixedKeys = make([]string, len(keys))
	)

	for i, key := range keys {
		prefixedKeys[i] = c.PrefixKey(key)
	}

	rst, err := c.client.MGet(ctx, prefixedKeys...).Result()
	if err != nil {
		return nil, err
	}

	for i, v := range rst {
		if v != nil {
			ret[keys[i]] = NewResult(v.(string), nil)
		} else {
			ret[keys[i]] = NewResult("", Nil)
		}
	}

	return ret, nil
}

// GetSet Retrieve or set an item from the cache by key.
func (c *RedisStore) GetSet(ctx context.Context, key string, fn defaultValueFunc) Result {
	var (
		prefixedKey = c.PrefixKey(key)
		cmd         = c.client.Get(ctx, prefixedKey)
	)

	if err := cmd.Err(); err != nil {
		if err != redis.Nil {
			return NewResult("", err)
		}

		switch ret, err := storeSharedCallGroup.Call(prefixedKey, func() (interface{}, error) {
			val, expire, err := fn()
			return defaultValueRet{
				val:    val,
				expire: expire,
			}, err
		}); err {
		case nil:
			ret := ret.(defaultValueRet)
			val := conv.String(ret.val)
			return NewResult(val, nil, c.Set(ctx, key, val, ret.expire))
		case Nil:
			ret := ret.(defaultValueRet)
			expire := c.GetDefaultNilExpire()
			if ret.expire > 0 {
				expire = ret.expire
			}
			return NewResult("", Nil, c.Set(ctx, key, c.GetDefaultNilValue(), expire))
		default:
			return NewResult("", err)
		}
	} else {
		if val := cmd.Val(); val == c.GetDefaultNilValue() {
			return NewResult("", Nil)
		} else {
			return NewResult(val)
		}
	}
}

// Set store an item in the cache for a given number of expire.
func (c *RedisStore) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, c.PrefixKey(key), conv.String(value), expiration).Err()
}

// SetMany store multiple items in the cache for a given number of expire.
func (c *RedisStore) SetMany(ctx context.Context, values map[string]interface{}, expiration time.Duration) error {
	var (
		lua  string
		keys = make([]string, 0, len(values))
		args = make([]interface{}, 1, len(values)+1)
	)

	if expiration > 0 {
		lua = `for i,k in ipairs(KEYS) do redis.call('SETEX',k,ARGV[1],ARGV[i+1]) end`
	} else {
		lua = `for i,k in ipairs(KEYS) do redis.call('SET',k,ARGV[i+1]) end`
	}

	args[1] = expiration / time.Second

	for key, value := range values {
		keys = append(keys, c.PrefixKey(key))
		args = append(args, conv.String(value))
	}

	return c.client.Eval(ctx, lua, keys, args...).Err()
}

// Forever store an item in the cache indefinitely.
func (c *RedisStore) Forever(ctx context.Context, key string, value interface{}) error {
	return c.client.Set(ctx, c.PrefixKey(key), value, redis.KeepTTL).Err()
}

// ForeverMany store multiple items in the cache indefinitely.
func (c *RedisStore) ForeverMany(ctx context.Context, values map[string]interface{}) error {
	return c.SetMany(ctx, values, redis.KeepTTL)
}

// Add store an item in the cache if the key does not exist.
func (c *RedisStore) Add(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, conv.String(value), expiration).Result()
}

// Increment increment the value of an item in the cache.
func (c *RedisStore) Increment(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, c.PrefixKey(key), value).Result()
}

// IncrementMany increment the value of multiple items in the cache.
func (c *RedisStore) IncrementMany(ctx context.Context, values map[string]int64) (map[string]int64, error) {
	var (
		lua          = `local r = {} for i,k in ipairs(KEYS) do table.insert(r, redis.call('INCRBY',k,ARGV[i])) end return r`
		count        = len(values)
		keys         = make([]string, 0, count)
		prefixedKeys = make([]string, 0, count)
		args         = make([]interface{}, 0, count)
	)

	for key, val := range values {
		prefixedKeys = append(prefixedKeys, c.PrefixKey(key))
		keys = append(keys, key)
		args = append(args, val)
	}

	rst, err := c.client.Eval(ctx, lua, prefixedKeys, args...).Result()
	if err != nil {
		return nil, err
	}

	ret := make(map[string]int64, count)

	for i, v := range rst.([]interface{}) {
		ret[keys[i]] = v.(int64)
	}

	return ret, nil
}

// Decrement decrement the value of an item in the cache.
func (c *RedisStore) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return c.Increment(ctx, key, value)
}

// DecrementMany decrement the value of multiple items in the cache.
func (c *RedisStore) DecrementMany(ctx context.Context, values map[string]int64) (map[string]int64, error) {
	return c.IncrementMany(ctx, values)
}

// Expire set expiration time for a key.
func (c *RedisStore) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.client.Expire(ctx, c.PrefixKey(key), expiration).Result()
}

// ExpireMany set expiration time for multiple key.
func (c *RedisStore) ExpireMany(ctx context.Context, values map[string]time.Duration) (map[string]bool, error) {
	var (
		lua          = `local r = {} for i,k in ipairs(KEYS) do table.insert(r, redis.call('EXPIRE',k,ARGV[i])) end return r`
		count        = len(values)
		keys         = make([]string, 0, count)
		prefixedKeys = make([]string, 0, count)
		args         = make([]interface{}, 0, count)
	)

	for key, expiration := range values {
		prefixedKeys = append(prefixedKeys, c.PrefixKey(key))
		keys = append(keys, key)
		args = append(args, expiration/time.Second)
	}

	rst, err := c.client.Eval(ctx, lua, prefixedKeys, args...).Result()
	if err != nil {
		return nil, err
	}

	ret := make(map[string]bool, count)

	for i, v := range rst.([]interface{}) {
		ret[keys[i]] = v.(int64) == 1
	}

	return ret, nil
}

// Forget remove an item from the cache.
func (c *RedisStore) Forget(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.PrefixKey(key)).Err()
}

// ForgetMany remove multiple items from the cache.
func (c *RedisStore) ForgetMany(ctx context.Context, keys ...string) (int64, error) {
	for i, key := range keys {
		keys[i] = c.PrefixKey(key)
	}

	return c.client.Del(ctx, keys...).Result()
}

// Flush remove all items from the cache.
func (c *RedisStore) Flush(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// Lock get a lock instance.
func (c *RedisStore) Lock(ctx context.Context, name string, time time.Duration) Lock {
	return NewRedisLock(c.client, ctx, c.PrefixKey(name), time)
}

// GetClient get the redis client instance.
func (c *RedisStore) GetClient() interface{} {
	return c.client
}
