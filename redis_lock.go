/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/5/23 4:13 下午
 * @Desc: a redis lock instance
 */

package cache

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"github.com/dobyte/cache/internal/safe"
)

type RedisLock struct {
	BaseLock
	client Redis
}

const (
	lockScript    = `if 1 == redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[2], 'NX') then return 1 end if ARGV[1] == redis.call('GET', KEYS[1]) then return redis.call('PEXPIRE', KEYS[1], ARGV[2]) end return 0`
	keepScript    = `if ARGV[1] == redis.call('GET', KEYS[1]) then redis.call('PEXPIRE', KEYS[1], ARGV[2]) return 1 end return 0`
	unlockScript  = `if ARGV[1] == redis.call('GET', KEYS[1]) then redis.call('DEL', KEYS[1]) return 1 end return 0`
	lockTicker    = 200 * time.Millisecond
	lockWatcher   = 10 * time.Second
	maxWatchTimes = 5
)

var (
	lockSha1   string
	keepSha1   string
	unlockSha1 string
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewRedisLock create a redis lock instance.
func NewRedisLock(client Redis, ctx context.Context, name string, time time.Duration) Lock {
	return &RedisLock{
		BaseLock: BaseLock{
			ctx:  ctx,
			code: strconv.FormatInt(rand.Int63(), 10),
			name: name,
			time: time,
			done: make(chan struct{}),
		},
		client: client,
	}
}

// Acquire attempt to acquire the lock.
func (l *RedisLock) Acquire() (ok bool, err error) {
	if ok, err = l.lock(); err != nil || ok {
		return
	}

	timer, ticker := time.NewTimer(l.time+lockTicker), time.NewTicker(lockTicker)

	for {
		select {
		case <-ticker.C:
			select {
			case <-timer.C:
				timer.Stop()
				ticker.Stop()
				return
			default:
				if ok, err = l.lock(); err != nil || ok {
					timer.Stop()
					ticker.Stop()
					return
				}
			}
		}
	}
}

// Release release the lock.
func (l *RedisLock) Release() (bool, error) {
	return l.unlock()
}

// lock and start watching goroutine to keep the lock valid.
func (l *RedisLock) lock() (ok bool, err error) {
	if lockSha1 == "" {
		if err = l.loadLockScript(); err != nil {
			return
		}
	}

	if ret, err := l.client.EvalSha(l.ctx, lockSha1, []string{l.name}, l.code, int64(l.time/time.Millisecond)).Int(); err != nil {
		safe.Go(func() {
			_ = l.loadLockScript()
		})
	} else if ok = ret == 1; ok {
		safe.Go(func() {
			l.watch()
		})
	}

	return
}

// unlock and destroy the watching goroutine.
func (l *RedisLock) unlock() (ok bool, err error) {
	if unlockSha1 == "" {
		if err = l.loadUnlockScript(); err != nil {
			return
		}
	}

	if ret, err := l.client.EvalSha(l.ctx, unlockSha1, []string{l.name}, l.code).Int(); err != nil {
		safe.Go(func() {
			_ = l.loadUnlockScript()
		})
	} else if ok = ret == 1; ok {
		l.done <- struct{}{}
	}

	return
}

// prevent the locker from being lost until the task is completed
func (l *RedisLock) keep() (ok bool, err error) {
	if keepSha1 == "" {
		if err = l.loadKeepScript(); err != nil {
			return
		}
	}

	if ret, err := l.client.EvalSha(l.ctx, keepSha1, []string{l.name}, l.code, int64(l.time/time.Millisecond)).Int(); err != nil {
		safe.Go(func() {
			_ = l.loadKeepScript()
		})
	} else {
		ok = ret == 1
	}

	return
}

// watch the locker, prevent the locker from being lost until the task is completed.
func (l *RedisLock) watch() {
	watcher := time.NewTicker(lockWatcher)

	for {
		select {
		case <-l.done:
			watcher.Stop()
			return
		case <-watcher.C:
			if ok, err := l.keep(); err != nil || !ok {
				watcher.Stop()
				return
			}

			l.times++

			if l.times >= maxWatchTimes {
				watcher.Stop()
				return
			}
		}
	}
}

// loads keep script.
func (l *RedisLock) loadKeepScript() (err error) {
	if sha1, err := l.client.ScriptLoad(l.ctx, keepScript).Result(); err == nil {
		keepSha1 = sha1
	}

	return
}

// loads lock script.
func (l *RedisLock) loadLockScript() (err error) {
	if sha1, err := l.client.ScriptLoad(l.ctx, lockScript).Result(); err == nil {
		lockSha1 = sha1
	}

	return
}

// loads unlock script.
func (l *RedisLock) loadUnlockScript() (err error) {
	if sha1, err := l.client.ScriptLoad(l.ctx, unlockScript).Result(); err == nil {
		unlockSha1 = sha1
	}

	return
}
