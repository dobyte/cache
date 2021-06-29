/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/5/23 4:10 下午
 * @Desc: lock interface define
 */

package cache

import "time"

type Lock interface {
	// Acquire Attempt to acquire the lock.
	Acquire() (bool, error)
	// Release Release the lock.
	Release() error
}

type BaseLock struct {
	name string
	time time.Duration
}
