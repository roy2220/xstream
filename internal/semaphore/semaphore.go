package semaphore

import (
	"context"
	"fmt"
	"sync/atomic"
	"unsafe"

	"github.com/roy2220/intrusive"
)

type Semaphore struct {
	availableN  int
	waiterList  intrusive.List
	closedError error
	isClosed    int32
}

func (s *Semaphore) Init(availableN int) *Semaphore {
	if availableN < 0 {
		panic(fmt.Errorf("semaphore: illegal argument; availableN=%d", availableN))
	}

	s.availableN = availableN
	s.waiterList.Init()
	return s
}

func (s *Semaphore) WaitFor(ctx context.Context, lock Lock, n int, callback func()) error {
	if n < 0 {
		panic(fmt.Errorf("semaphore: illegal argument; n=%d", n))
	}

	waiter, err := s.tryWaitForOrEngageWaiter(ctx, lock, n, callback)

	if err != nil {
		return err
	}

	if waiter == nil {
		return nil
	}

	if err := waiter.GetNotification(ctx); err != nil {
		ok, err2 := s.dismissWaiter(lock, waiter)

		if err2 != nil {
			return err2
		}

		if !ok {
			return nil
		}

		return err
	}

	if atomic.LoadInt32(&s.isClosed) != 0 {
		return s.closedError
	}

	return nil
}

func (s *Semaphore) TryWaitFor(ctx context.Context, lock Lock, n int, callback func()) (bool, error) {
	if n < 0 {
		panic(fmt.Errorf("semaphore: illegal argument; n=%d", n))
	}

	if err := lock.Acquire(ctx); err != nil {
		return false, err
	}

	defer lock.Release()
	return s.tryWaitForWithoutLock(n, callback), nil
}

func (s *Semaphore) Signal(ctx context.Context, lock Lock, n int) error {
	if n < 0 {
		panic(fmt.Errorf("semaphore: illegal argument; n=%d", n))
	}

	if err := lock.Acquire(ctx); err != nil {
		return err
	}

	defer lock.Release()
	s.availableN += n
	s.notifyWaiters()
	return nil
}

func (s *Semaphore) Close(closedError error) {
	s.closedError = closedError
	atomic.StoreInt32(&s.isClosed, 1)

	for it := s.waiterList.Foreach(); !it.IsAtEnd(); it.Advance() {
		waiter := getWaiter(it.Node())
		waiter.ListNode = intrusive.ListNode{}
		waiter.Notify()
	}

	s.waiterList = intrusive.List{}
}

func (s *Semaphore) tryWaitForOrEngageWaiter(ctx context.Context, lock Lock, n int, callback func()) (*waiter, error) {
	if err := lock.Acquire(ctx); err != nil {
		return nil, err
	}

	defer lock.Release()

	if s.tryWaitForWithoutLock(n, callback) {
		return nil, nil
	}

	return s.engageWaiterWithoutLock(n, callback), nil
}

func (s *Semaphore) tryWaitForWithoutLock(n int, callback func()) bool {
	if !s.waiterList.IsEmpty() || n > s.availableN {
		return false
	}

	s.availableN -= n
	callback()
	return true
}

func (s *Semaphore) engageWaiterWithoutLock(n int, callback func()) *waiter {
	waiter := (&waiter{
		N:        n,
		Callback: callback,
	}).Init()

	s.waiterList.AppendNode(&waiter.ListNode)
	return waiter
}

func (s *Semaphore) dismissWaiter(lock Lock, waiter *waiter) (bool, error) {
	if err := lock.Acquire(context.Background()); err != nil {
		return false, err
	}

	defer lock.Release()

	if waiter.ListNode.IsReset() {
		return false, nil
	}

	waiterWasFront := &waiter.ListNode == s.waiterList.Head()
	waiter.ListNode.Remove()
	waiter.ListNode = intrusive.ListNode{}

	if waiterWasFront {
		s.notifyWaiters()
	}

	return true, nil
}

func (s *Semaphore) notifyWaiters() {
	for it := s.waiterList.Foreach(); !it.IsAtEnd(); it.Advance() {
		waiter := getWaiter(it.Node())

		if waiter.N > s.availableN {
			return
		}

		s.availableN -= waiter.N
		waiter.Callback()
		waiter.ListNode.Remove()
		waiter.ListNode = intrusive.ListNode{}
		waiter.Notify()
	}
}

type waiter struct {
	ListNode intrusive.ListNode
	N        int
	Callback func()

	notification chan struct{}
}

func (w *waiter) Init() *waiter {
	w.notification = make(chan struct{})
	return w
}

func (w *waiter) GetNotification(ctx context.Context) error {
	select {
	case <-w.notification:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *waiter) Notify() {
	close(w.notification)
}

func getWaiter(listNode *intrusive.ListNode) *waiter {
	return (*waiter)(listNode.GetContainer(unsafe.Offsetof(waiter{}.ListNode)))
}
