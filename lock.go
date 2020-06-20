package xstream

import (
	"context"

	"errors"
	"github.com/roy2220/xstream/internal/flowcontroller"
)

type lock struct {
	c chan struct{}
}

var _ = flowcontroller.Lock(&lock{})

func (l *lock) Init() *lock {
	c := make(chan struct{}, 1)
	c <- struct{}{}
	l.c = c
	return l
}

func (l *lock) Acquire(ctx context.Context) error {
	select {
	case _, ok := <-l.c:
		if !ok {
			return l.ClosedError()
		}

		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *lock) Release() {
	select {
	case l.c <- struct{}{}:
	default:
		panic(errors.New("xstream: lock not acquired"))
	}
}

func (l *lock) Close() error {
	if _, ok := <-l.c; !ok {
		return l.ClosedError()
	}

	close(l.c)
	return nil
}

func (l *lock) ClosedError() error {
	return ErrClosed
}
