/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/6/2 9:44 上午
 * @Desc: a memcached lock instance
 */

package cache

import (
	"time"
	
	"github.com/bradfitz/gomemcache/memcache"
)

type MemcachedLock struct {
	BaseLock
	client *Memcached
}

// NewMemcachedLock Create a memcached lock instance.
func NewMemcachedLock(client *Memcached, name string, time time.Duration) Lock {
	return &MemcachedLock{
		BaseLock: BaseLock{
			name: name,
			time: time,
		},
		client: client,
	}
}

// Acquire Attempt to acquire the lock.
func (l *MemcachedLock) Acquire() (bool, error) {
	if err := l.client.Add(&memcache.Item{
		Key:        l.name,
		Value:      []byte("1"),
		Expiration: int32(l.time / time.Second),
	}); err != nil {
		if err == memcache.ErrNotStored {
			return false, nil
		}
		
		return false, err
	} else {
		return true, nil
	}
}

// Release Release the lock.
func (l *MemcachedLock) Release() error {
	if err := l.client.Delete(l.name); err != nil && err != memcache.ErrCacheMiss {
		return err
	}
	
	return nil
}
