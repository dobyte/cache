/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/5/23 4:10 下午
 * @Desc: lock interface define
 */

package cache

import (
	"context"
	"time"
)

type Lock interface {
	// Acquire attempt to acquire the lock.
	Acquire() (bool, error)
	// Release release the lock.
	Release() (bool, error)
}

type BaseLock struct {
	ctx   context.Context
	code  string
	name  string
	time  time.Duration
	done  chan struct{}
	times int
}
