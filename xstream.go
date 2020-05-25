package xstream

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/roy2220/xstream/internal/flowcontroller"
	"github.com/roy2220/xstream/internal/stream"
)

type XStream struct {
	size           int64
	flowController flowcontroller.FlowController
	s              stream.Stream
	lock           lock
}

func (xs *XStream) Init(size int) *XStream {
	xs.size = int64(size)
	xs.flowController.Init(size, &xs.lock)
	return xs
}

func (xs *XStream) ReadFull(ctx context.Context, dataSize int, callback func(data []byte)) error {
	if dataSize > xs.Size() {
		return ErrSizeExceeded
	}

	return xs.flowController.ReceiveData(ctx, dataSize, func() {
		data := xs.s.Data()[:dataSize]
		callback(data)
		xs.s.DiscardData(dataSize)
	})
}

func (xs *XStream) TryReadFull(dataSize int, callback func(data []byte)) (bool, error) {
	ok, err := xs.flowController.TryReceiveData(dataSize, func() {
		data := xs.s.Data()[:dataSize]
		callback(data)
		xs.s.DiscardData(dataSize)
	})

	return ok, err
}

func (xs *XStream) WriteAll(ctx context.Context, bufferSize int, callback func(buffer []byte)) error {
	if bufferSize > xs.Size() {
		return ErrSizeExceeded
	}

	return xs.flowController.SendData(ctx, bufferSize, func() {
		buffer := xs.s.AcquireBuffer(bufferSize)
		callback(buffer)
		xs.s.CommitData(bufferSize)
	})
}

func (xs *XStream) TryWriteAll(bufferSize int, callback func(buffer []byte)) (bool, error) {
	ok, err := xs.flowController.TrySendData(bufferSize, func() {
		buffer := xs.s.AcquireBuffer(bufferSize)
		callback(buffer)
		xs.s.CommitData(bufferSize)
	})

	return ok, err
}

func (xs *XStream) Enlarge(additionalSize int) error {
	if err := xs.flowController.GrowBuffer(additionalSize); err != nil {
		return err
	}

	atomic.AddInt64(&xs.size, int64(additionalSize))
	return nil
}

func (xs *XStream) Peek(callback func(data []byte)) error {
	if err := xs.lock.Acquire(); err != nil {
		return err
	}

	defer xs.lock.Release()
	callback(xs.s.Data())
	return nil
}

func (xs *XStream) Close() error {
	return xs.flowController.Close()
}

func (xs *XStream) Size() int {
	return int(atomic.LoadInt64(&xs.size))
}

var (
	ErrClosed       = errors.New("xstream: closed")
	ErrSizeExceeded = errors.New("xstream: size exceeded")
)
