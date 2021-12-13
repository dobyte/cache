/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/6/5 7:24 下午
 * @Desc: TODO
 */

package cache_test

import (
	"context"
	"testing"

	"github.com/dobyte/cache"
)

func newRedisCache() cache.Cache {
	return cache.NewCache(&cache.Options{
		Driver: cache.RedisDriver,
		Prefix: "cache",
		Stores: cache.Stores{
			Redis: &cache.RedisOptions{
				Addrs: []string{"127.0.0.1:6379"},
			},
		},
	})
}

func newMemcachedCache() cache.Cache {
	return cache.NewCache(&cache.Options{
		Driver: cache.MemcachedDriver,
		Prefix: "cache",
		Stores: cache.Stores{
			Memcached: &cache.MemcachedOptions{
				Addrs: []string{":11211"},
			},
		},
	})
}

// func TestCache_Has(t *testing.T) {
// 	var (
// 		redis     = newRedisCache()
// 		memcached = newMemcachedCache()
// 		err       error
// 		key       = "name"
// 	)
//
// 	_, err = redis.Has(key)
// 	if err != nil {
// 		t.Fatalf("redis: failed to detect the existence of cache: %v", err.Error())
// 		return
// 	}
//
// 	_, err = memcached.Has(key)
// 	if err != nil {
// 		t.Fatalf("mc: failed to detect the existence of cache: %v", err.Error())
// 		return
// 	}
// }
//
// func TestCache_Get(t *testing.T) {
// 	var (
// 		redis        = newRedisCache()
// 		memcached    = newMemcachedCache()
// 		err          error
// 		val          string
// 		key          = "name"
// 		defaultValue = "fuxiao"
// 	)
//
// 	err = redis.Get(key).Err()
// 	if err != nil && err != cache.Nil {
// 		t.Fatalf("redis: failed to retrieve an item from cache: %v", err.Error())
// 	}
//
// 	err = memcached.Get(key).Err()
// 	if err != nil && err != cache.Nil {
// 		t.Fatalf("mc: failed to retrieve an item from cache: %v", err.Error())
// 	}
//
// 	val, err = redis.Get(key, defaultValue).Result()
// 	if err != nil {
// 		t.Fatalf("redis: failed to retrieve an item from cache: %v", err.Error())
// 	} else {
// 		t.Log(val)
// 	}
//
// 	val, err = memcached.Get(key, defaultValue).Result()
// 	if err != nil {
// 		t.Fatalf("mc: failed to retrieve an item from cache: %v", err.Error())
// 	} else {
// 		t.Log(val)
// 	}
// }
//
func TestCache_GetMany(t *testing.T) {
	redis := newRedisCache()

	ret, err := redis.GetMany(context.Background(), "a", "b")
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range ret {
		t.Log(k)
		t.Log(v.Result())
	}
}

func TestCache_HasMany(t *testing.T) {
	redis := newRedisCache()

	ret, err := redis.HasMany(context.Background(), "a", "b")
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range ret {
		t.Log(k)
		t.Log(v)
	}
}

//
// func TestCache_GetSet(t *testing.T) {
// 	var (
// 		redis         = newRedisCache()
// 		memcached     = newMemcachedCache()
// 		err           error
// 		val           string
// 		key           = "name"
// 		defaultValue  = "fuxiao"
// 		defaultExpire = 10 * time.Second
// 	)
//
// 	val, err = redis.GetSet(key, func() (interface{}, time.Duration, error) {
// 		t.Log("redis: reading resource...")
//
// 		return defaultValue, defaultExpire, nil
// 	}).Result()
// 	if err != nil && err != cache.Nil {
// 		t.Fatalf("redis: failed to retrieve an item from cache: %v", err.Error())
// 	} else {
// 		t.Logf("redis: read data from cache: %v", val)
// 	}
//
// 	val, err = memcached.GetSet(key, func() (interface{}, time.Duration, error) {
// 		t.Log("mc: reading resource...")
//
// 		return defaultValue, defaultExpire, nil
// 	}).Result()
// 	if err != nil && err != cache.Nil {
// 		t.Fatalf("mc: failed to retrieve an item from cache: %v", err.Error())
// 	} else {
// 		t.Logf("mc: read data from cache: %v", val)
// 	}
// }
//
// func TestCache_GetSet_ReadResourceFailed(t *testing.T) {
// 	var (
// 		redis           = newRedisCache()
// 		memcached       = newMemcachedCache()
// 		err             error
// 		val             string
// 		key             = "name"
// 		readResourceErr = fmt.Errorf("failed to read resource")
// 	)
//
// 	val, err = redis.GetSet(key, func() (interface{}, time.Duration, error) {
// 		return "", 0, readResourceErr
// 	}).Result()
// 	if err != nil && err != cache.Nil {
// 		if err == readResourceErr {
// 			t.Log("redis: " + readResourceErr.Error())
// 		} else {
// 			t.Fatalf("redis: failed to retrieve an item from cache: %v", err.Error())
// 		}
// 	} else {
// 		t.Logf("redis: read data from cache: %v", val)
// 	}
//
// 	val, err = memcached.GetSet(key, func() (interface{}, time.Duration, error) {
// 		return "", 0, readResourceErr
// 	}).Result()
// 	if err != nil && err != cache.Nil {
// 		if err == readResourceErr {
// 			t.Log("mc: " + readResourceErr.Error())
// 		} else {
// 			t.Fatalf("mc: failed to retrieve an item from cache: %v", err.Error())
// 		}
// 	} else {
// 		t.Logf("mc: read data from cache: %v", val)
// 	}
// }
//
// func TestCache_GetSet_ReadEmptyResource(t *testing.T) {
// 	var (
// 		redis     = newRedisCache()
// 		memcached = newMemcachedCache()
// 		err       error
// 		key       = "name"
// 	)
//
// 	_, err = redis.GetSet(key, func() (interface{}, time.Duration, error) {
// 		t.Log("redis: readed empty resource")
//
// 		return "", 0, cache.Nil
// 	}).Result()
// 	if err != nil && err != cache.Nil {
// 		t.Fatalf("redis: failed to retrieve an item from cache: %v", err.Error())
// 	} else {
// 		t.Logf("redis: read empty data from cache,and keep empty data for %v seconds", 10)
// 	}
//
// 	_, err = memcached.GetSet(key, func() (interface{}, time.Duration, error) {
// 		t.Log("mc: readed empty resource")
//
// 		return "", 0, cache.Nil
// 	}).Result()
// 	if err != nil && err != cache.Nil {
// 		t.Fatalf("mc: failed to retrieve an item from cache: %v", err.Error())
// 	} else {
// 		t.Logf("mc: readed empty data from cache,and keep empty data for %v seconds", 10)
// 	}
// }
//
// func TestCache_GetSet_SharedCallGroup(t *testing.T) {
// 	var (
// 		redis              = newRedisCache()
// 		memcached          = newMemcachedCache()
// 		err                error
// 		val                string
// 		redisKey           = "redis:name"
// 		mcKey              = "mc:name"
// 		redisDefaultValue  = "redis-fuxiao"
// 		redisDefaultExpire = 10 * time.Second
// 		mcDefaultValue     = "mc-fuxiao"
// 		mcDefaultExpire    = 10 * time.Second
// 		wg                 sync.WaitGroup
// 	)
//
// 	wg.Add(6)
//
// 	for i := 0; i < 3; i++ {
// 		go func() {
// 			val, err = redis.GetSet(redisKey, func() (interface{}, time.Duration, error) {
// 				t.Log("redis: reading resource...")
//
// 				time.Sleep(3 * time.Second)
//
// 				return redisDefaultValue, redisDefaultExpire, nil
// 			}).Result()
// 			if err != nil && err != cache.Nil {
// 				t.Fatalf("redis: failed to retrieve an item from cache: %v", err.Error())
// 			} else {
// 				t.Logf("redis: read data from cache: %v", val)
// 			}
//
// 			wg.Done()
// 		}()
// 	}
//
// 	for i := 0; i < 3; i++ {
// 		go func() {
// 			val, err = memcached.GetSet(mcKey, func() (interface{}, time.Duration, error) {
// 				t.Log("mc: reading resource...")
//
// 				time.Sleep(3 * time.Second)
//
// 				return mcDefaultValue, mcDefaultExpire, nil
// 			}).Result()
// 			if err != nil && err != cache.Nil {
// 				t.Fatalf("mc: failed to retrieve an item from cache: %v", err.Error())
// 			} else {
// 				t.Logf("mc: read data from cache: %v", val)
// 			}
//
// 			wg.Done()
// 		}()
// 	}
//
// 	wg.Wait()
// }
//
// func TestCache_Forget(t *testing.T) {
// 	var (
// 		redis     = newRedisCache()
// 		memcached = newMemcachedCache()
// 		err       error
// 		key       = "name"
// 		value     = "fuxiao"
// 	)
//
// 	err = redis.Set(key, value, 0)
// 	if err != nil {
// 		t.Fatalf("redis: failed to store data to the cache: %v", err.Error())
// 	}
//
// 	err = memcached.Set(key, value, 0)
// 	if err != nil {
// 		t.Fatalf("mc: failed to store data to the cache: %v", err.Error())
// 	}
//
// 	err = redis.Forget(key)
// 	if err != nil {
// 		t.Fatalf("redis: failed to delete data from cache: %v", err.Error())
// 	}
//
// 	err = memcached.Forget(key)
// 	if err != nil {
// 		t.Fatalf("mc: failed to delete data from cache: %v", err.Error())
// 	}
// }
//
// func TestCache_Flush(t *testing.T) {
// 	var (
// 		redis     = newRedisCache()
// 		memcached = newMemcachedCache()
// 		err       error
// 	)
//
// 	err = redis.Flush()
// 	if err != nil {
// 		t.Fatalf("redis: failed to clean all cache: %v", err.Error())
// 	}
//
// 	err = memcached.Flush()
// 	if err != nil {
// 		t.Fatalf("mc: failed to clean all cache: %v", err.Error())
// 	}
// }
//
// func TestCache_GetClient(t *testing.T) {
// 	// var (
// 	// 	redis     = newRedisCache()
// 	// 	memcached = newMemcachedCache()
// 	// 	err       error
// 	// )
// 	//
// 	// _, err = redis.GetClient().(*cache.Redis).Ping(context.Background()).Result()
// 	// if err != nil {
// 	// 	t.Fatalf("redis: failed to send ping command with native client: %v", err.Error())
// 	// }
// 	//
// 	// err = memcached.GetClient().(*cache.Memcached).Ping()
// 	// if err != nil {
// 	// 	t.Fatalf("mc: failed to send ping command with native client: %v", err.Error())
// 	// }
// }
