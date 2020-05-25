package xstream

import (
	"sync"
	"sync/atomic"

	"github.com/roy2220/xstream/internal/flowcontroller"
)

type lock struct {
	isClosed int32
	mutex    sync.Mutex
}

var _ = flowcontroller.Lock(&lock{})

func (l *lock) Acquire() error {
	if atomic.LoadInt32(&l.isClosed) == 1 {
		return l.ClosedError()
	}

	l.mutex.Lock()

	if l.isClosed == 1 {
		l.mutex.Unlock()
		return l.ClosedError()
	}

	return nil
}

func (l *lock) Release() {
	l.mutex.Unlock()
}

func (l *lock) Close() error {
	if err := l.Acquire(); err != nil {
		return err
	}

	defer l.Release()
	atomic.StoreInt32(&l.isClosed, 1)
	return nil
}

func (l *lock) ClosedError() error {
	return ErrClosed
}
