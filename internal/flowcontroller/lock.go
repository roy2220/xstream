package flowcontroller

import "github.com/roy2220/xstream/internal/semaphore"

type Lock interface {
	semaphore.Lock

	Close() (err error)
	ClosedError() (err error)
}
