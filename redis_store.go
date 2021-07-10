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

const (
	RedisNodeMode    = "node"
	RedisClusterMode = "cluster"
)

type (
	Redis      = redis.Cmdable
	RedisStore struct {
		BaseStore
		client Redis
	}
	
	RedisOptions struct {
		Mode             string
		Addrs            []string
		Username         string
		Password         string
		DB               int
		Prefix           string
		DefaultNilValue  string
		DefaultNilExpire time.Duration
	}
)

// NewRedisStore Create a redis store instance.
func NewRedisStore(opt *RedisOptions) Store {
	var client redis.Cmdable
	
	if opt.Mode == RedisClusterMode {
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    opt.Addrs,
			Username: opt.Username,
			Password: opt.Password,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:     opt.Addrs[0],
			Username: opt.Username,
			Password: opt.Password,
			DB:       opt.DB,
		})
	}
	
	c := &RedisStore{client: client}
	c.SetPrefix(opt.Prefix)
	c.SetDefaultNilValue(opt.DefaultNilValue)
	c.SetDefaultNilExpire(opt.DefaultNilExpire)
	
	return c
}

// Get Retrieve an item from the cache by key.
func (c *RedisStore) Get(key string, defaultValue ...interface{}) Result {
	if val, err := c.client.Get(context.Background(), c.PrefixKey(key)).Result(); err == redis.Nil {
		if len(defaultValue) > 0 {
			return NewResult(conv.String(defaultValue[0]))
		} else {
			return NewResult("", Nil)
		}
	} else {
		return NewResult(val, err)
	}
}

// GetMany Retrieve multiple items from the cache by key.
func (c *RedisStore) GetMany(keys ...string) (map[string]string, error) {
	var (
		ctx  = context.Background()
		pipe = c.client.Pipeline()
		ret  = make(map[string]string)
	)
	
	for _, key := range keys {
		ret[key] = pipe.Get(ctx, c.PrefixKey(key)).Val()
	}
	
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}
	
	return ret, nil
}

// GetSet Retrieve or set an item from the cache by key.
func (c *RedisStore) GetSet(key string, fn defaultValueFunc) Result {
	var (
		prefixedKey = c.PrefixKey(key)
		cmd         = c.client.Get(context.Background(), prefixedKey)
	)
	
	if err := cmd.Err(); err != nil {
		if err != redis.Nil {
			return NewResult("", err)
		} else {
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
		if val := cmd.Val(); val == c.GetDefaultNilValue() {
			return NewResult("", Nil)
		} else {
			return NewResult(val)
		}
	}
}

// Has Determine if an items exists in the cache.
func (c *RedisStore) Has(key string) (bool, error) {
	ret := c.client.Exists(context.Background(), c.PrefixKey(key))
	
	return ret.Val() != 0, ret.Err()
}

// HasMany Determine if multiple item exists in the cache.
func (c *RedisStore) HasMany(keys ...string) (map[string]bool, error) {
	var (
		ctx  = context.Background()
		pipe = c.client.Pipeline()
		ret  = make(map[string]bool)
	)
	
	for _, key := range keys {
		ret[key] = pipe.Exists(ctx, c.PrefixKey(key)).Val() != 0
	}
	
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}
	
	return ret, nil
}

// Set Store an item in the cache for a given number of expire.
func (c *RedisStore) Set(key string, value interface{}, expire time.Duration) error {
	return c.client.Set(context.Background(), c.PrefixKey(key), conv.String(value), expire).Err()
}

// SetMany Store multiple items in the cache for a given number of expire.
func (c *RedisStore) SetMany(values map[string]interface{}, expire time.Duration) error {
	var (
		ctx  = context.Background()
		pipe = c.client.Pipeline()
	)
	
	for key, value := range values {
		pipe.Set(ctx, c.PrefixKey(key), conv.String(value), expire)
	}
	
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	
	return nil
}

// Forever Store an item in the cache indefinitely.
func (c *RedisStore) Forever(key string, value interface{}) error {
	return c.client.Set(context.Background(), c.PrefixKey(key), value, 0).Err()
}

// ForeverMany Store multiple items in the cache indefinitely.
func (c *RedisStore) ForeverMany(values map[string]interface{}) error {
	var (
		ctx  = context.Background()
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
func (c *RedisStore) Add(key string, value interface{}, expire time.Duration) (bool, error) {
	lua := "return redis.call('exists',KEYS[1])<1 and redis.call('setex',KEYS[1],ARGV[2],ARGV[1])"
	
	return c.client.Eval(context.Background(), lua, []string{c.PrefixKey(key)}, value, expire/time.Second).Bool()
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
