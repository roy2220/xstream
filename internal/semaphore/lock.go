package semaphore

type Lock interface {
	Acquire() error
	Release()
}

var NoLock = noLock{}

type noLock struct{}

var _ = Lock(noLock{})

func (noLock) Acquire() error { return nil }
func (noLock) Release()       {}
