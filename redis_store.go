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

// Set Store an item in the cache for a given number of expire.
func (c *RedisStore) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, c.PrefixKey(key), conv.String(value), expiration).Err()
}

// SetMany Store multiple items in the cache for a given number of expire.
func (c *RedisStore) SetMany(ctx context.Context, values map[string]interface{}, expiration time.Duration) error {
	var (
		lua          = `for i,k in ipairs(KEYS) do redis.call('setex',k,ARGV[1],ARGV[i+1]) end`
		prefixedKeys = make([]string, 0, len(values))
		args         = make([]interface{}, 1, len(values)+1)
	)

	args[1] = expiration / time.Second

	for key, value := range values {
		prefixedKeys = append(prefixedKeys, c.PrefixKey(key))
		args = append(args, conv.String(value))
	}
	c.client.MGet()
	c.client.MSet()

	c.client.PExpire()

	return c.client.Eval(ctx, lua, prefixedKeys, args...).Err()
}

// Forever Store an item in the cache indefinitely.
func (c *RedisStore) Forever(ctx context.Context, key string, value interface{}) error {
	return c.client.Set(ctx, c.PrefixKey(key), value, 0).Err()
}

// ForeverMany Store multiple items in the cache indefinitely.
func (c *RedisStore) ForeverMany(ctx context.Context, values map[string]interface{}) error {
	var (
		pipe = c.client.Pipeline()
	)

	for key, value := range values {
		pipe.Set(ctx, c.PrefixKey(key), value, 0)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return nil
}

// Add Store an item in the cache if the key does not exist.
func (c *RedisStore) Add(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if expiration > 0 {

	} else if expiration == redis.KeepTTL {

	}

	lua := "return redis.call('exists',KEYS[1])<1 and redis.call('setex',KEYS[1],ARGV[2],ARGV[1])"

	return c.client.Eval(ctx, lua, []string{c.PrefixKey(key)}, value, expiration/time.Second).Bool()
}

// Increment Increment the value of an item in the cache.
func (c *RedisStore) Increment(key string, value int64) (int64, error) {
	return c.client.IncrBy(context.Background(), c.PrefixKey(key), value).Result()
}

// IncrementMany Increment the value of multiple items in the cache.
func (c *RedisStore) IncrementMany(values map[string]int64) (map[string]int64, error) {
	var (
		ctx  = context.Background()
		pipe = c.client.Pipeline()
		ret  = make(map[string]int64)
	)

	for key, value := range values {
		ret[key] = pipe.IncrBy(ctx, c.PrefixKey(key), value).Val()
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	return ret, nil
}

// Decrement Decrement the value of an item in the cache.
func (c *RedisStore) Decrement(key string, value int64) (int64, error) {
	ret := c.client.DecrBy(context.Background(), c.PrefixKey(key), value)

	return ret.Val(), ret.Err()
}

// DecrementMany Decrement the value of multiple items in the cache.
func (c *RedisStore) DecrementMany(values map[string]int64) (map[string]int64, error) {
	var (
		ctx  = context.Background()
		pipe = c.client.Pipeline()
		ret  = make(map[string]int64)
	)

	for key, value := range values {
		ret[key] = pipe.DecrBy(ctx, c.PrefixKey(key), value).Val()
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	return ret, nil
}

// Expire Set expiration time for a key.
func (c *RedisStore) Expire(key string, expire time.Duration) (bool, error) {
	return c.client.Expire(context.Background(), c.PrefixKey(key), expire).Result()
}

// ExpireMany Expire Set expiration time for multiple key.
func (c *RedisStore) ExpireMany(values map[string]time.Duration) (map[string]bool, error) {
	var (
		ctx  = context.Background()
		pipe = c.client.Pipeline()
		ret  = make(map[string]bool)
	)

	for key, expire := range values {
		ret[key] = pipe.Expire(ctx, c.PrefixKey(key), expire).Val()
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	return ret, nil
}

// Forget Remove an item from the cache.
func (c *RedisStore) Forget(key string) error {
	return c.client.Del(context.Background(), c.PrefixKey(key)).Err()
}

// ForgetMany Remove multiple items from the cache.
func (c *RedisStore) ForgetMany(keys ...string) (int64, error) {
	for i, key := range keys {
		keys[i] = c.PrefixKey(key)
	}

	ret := c.client.Del(context.Background(), keys...)

	return ret.Val(), ret.Err()
}

// Flush Remove all items from the cache.
func (c *RedisStore) Flush() error {
	return c.client.FlushDB(context.Background()).Err()
}

// Lock Get a lock instance.
func (c *RedisStore) Lock(name string, time time.Duration) Lock {
	return NewRedisLock(c.client, c.PrefixKey(name), time)
}

// GetClient Get the redis client instance.
func (c *RedisStore) GetClient() interface{} {
	return c.client
}
