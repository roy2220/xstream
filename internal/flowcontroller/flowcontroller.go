package flowcontroller

import (
	"context"

	"github.com/roy2220/xstream/internal/semaphore"
)

type FlowController struct {
	dataSemaphore   semaphore.Semaphore
	bufferSemaphore semaphore.Semaphore
	lock            Lock
}

func (fc *FlowController) Init(bufferSize int, lock Lock) *FlowController {
	fc.dataSemaphore.Init(0)
	fc.bufferSemaphore.Init(bufferSize)
	fc.lock = lock
	return fc
}

func (fc *FlowController) Close() error {
	if err := fc.lock.Close(); err != nil {
		return err
	}

	closedError := fc.lock.ClosedError()
	fc.dataSemaphore.Close(closedError)
	fc.bufferSemaphore.Close(closedError)
	return nil
}

func (fc *FlowController) ReceiveData(ctx context.Context, dataSize int, callback func()) error {
	return fc.dataSemaphore.WaitFor(ctx, fc.lock, dataSize, func() {
		callback()
		fc.bufferSemaphore.Signal(context.Background(), semaphore.NoLock, dataSize)
	})
}

func (fc *FlowController) TryReceiveData(ctx context.Context, dataSize int, callback func()) (bool, error) {
	return fc.dataSemaphore.TryWaitFor(ctx, fc.lock, dataSize, func() {
		callback()
		fc.bufferSemaphore.Signal(context.Background(), semaphore.NoLock, dataSize)
	})
}

func (fc *FlowController) SendData(ctx context.Context, dataSize int, callback func()) error {
	return fc.bufferSemaphore.WaitFor(ctx, fc.lock, dataSize, func() {
		callback()
		fc.dataSemaphore.Signal(context.Background(), semaphore.NoLock, dataSize)
	})
}

func (fc *FlowController) TrySendData(ctx context.Context, dataSize int, callback func()) (bool, error) {
	return fc.bufferSemaphore.TryWaitFor(ctx, fc.lock, dataSize, func() {
		callback()
		fc.dataSemaphore.Signal(context.Background(), semaphore.NoLock, dataSize)
	})
}

func (fc *FlowController) GrowBuffer(ctx context.Context, additionalBufferSize int) error {
	return fc.bufferSemaphore.Signal(ctx, fc.lock, additionalBufferSize)
}
