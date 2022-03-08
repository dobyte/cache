/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/5/21 10:01 上午
 * @Desc: a shared call group instance
 */

package sync

import "sync"

type (
	sharedCall struct {
		wg  sync.WaitGroup
		val interface{}
		err error
	}

	sharedCallGroup struct {
		calls  map[string]*sharedCall
		locker sync.Mutex
	}

	CallFunc func() (interface{}, error)
)

func NewSharedCallGroup() *sharedCallGroup {
	return &sharedCallGroup{
		calls: make(map[string]*sharedCall),
	}
}

func (s *sharedCallGroup) Call(key string, fn func() (interface{}, error)) (interface{}, error) {
	call, done := s.createCall(key)
	if done {
		return call.val, call.err
	}

	s.makeCall(key, call, fn)

	return call.val, call.err
}

func (s *sharedCallGroup) createCall(key string) (*sharedCall, bool) {
	s.locker.Lock()

	if call, ok := s.calls[key]; ok {
		s.locker.Unlock()
		call.wg.Wait()
		return call, true
	}

	call := new(sharedCall)
	s.calls[key] = call
	call.wg.Add(1)
	s.locker.Unlock()
	return call, false
}

func (s *sharedCallGroup) makeCall(key string, call *sharedCall, fn func() (interface{}, error)) {
	defer func() {
		s.locker.Lock()
		delete(s.calls, key)
		s.locker.Unlock()
		call.wg.Done()
	}()

	call.val, call.err = fn()
}
