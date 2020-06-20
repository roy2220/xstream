package semaphore

import "context"

type Lock interface {
	Acquire(ctx context.Context) (err error)
	Release()
}

var NoLock = noLock{}

type noLock struct{}

var _ = Lock(noLock{})

func (noLock) Acquire(ctx context.Context) (err error) { return }
func (noLock) Release()                                {}
