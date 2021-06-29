/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/5/23 4:13 下午
 * @Desc: a redis lock instance
 */

package cache

import (
	"context"
	"time"
)

type RedisLock struct {
	BaseLock
	client Redis
}

// NewRedisLock Create a redis lock instance.
func NewRedisLock(client Redis, name string, time time.Duration) Lock {
	return &RedisLock{
		BaseLock: BaseLock{
			name: name,
			time: time,
		},
		client: client,
	}
}

// Acquire Attempt to acquire the lock.
func (l *RedisLock) Acquire() (bool, error) {
	return l.client.SetNX(context.Background(), l.name, 1, l.time).Result()
}

// Release Release the lock.
func (l *RedisLock) Release() error {
	return l.client.Del(context.Background(), l.name).Err()
}
