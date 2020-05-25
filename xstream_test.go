package xstream_test

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/roy2220/xstream"
	"github.com/stretchr/testify/assert"
)

func TestStreamReadFull(t *testing.T) {
	s := new(xstream.XStream).Init(100)

	err := s.ReadFull(context.Background(), 0, func(data []byte) {
		assert.Len(t, data, 0)
	})
	assert.NoError(t, err)

	err = s.ReadFull(context.Background(), 101, func([]byte) {})
	assert.Error(t, err, xstream.ErrSizeExceeded)

	go func() {
		ok, err := s.TryWriteAll(25, func(buffer []byte) {
			assert.Len(t, buffer, 25)
		})
		assert.True(t, ok)
		assert.NoError(t, err)
	}()

	for i := 0; i < 2; i++ {
		err = s.ReadFull(context.Background(), 10, func(data []byte) {
			assert.Len(t, data, 10)
		})
		assert.NoError(t, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err = s.ReadFull(ctx, 10, func(data []byte) {
		assert.Len(t, data, 10)
	})
	assert.Error(t, err, context.DeadlineExceeded)

	go func() {
		time.Sleep(50 * time.Millisecond)
		ok, err := s.TryWriteAll(5, func(buffer []byte) {
			assert.Len(t, buffer, 5)
		})
		assert.True(t, ok)
		assert.NoError(t, err)
	}()

	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err = s.ReadFull(ctx, 10, func(data []byte) {
		assert.Len(t, data, 10)
	})
	assert.NoError(t, err)

	go func() {
		err := s.Close()
		assert.NoError(t, err)
	}()

	err = s.ReadFull(context.Background(), 10, func([]byte) {})
	assert.Error(t, err, xstream.ErrClosed)
	err = s.ReadFull(context.Background(), 0, func([]byte) {})
	assert.Error(t, err, xstream.ErrClosed)
}

func TestStreamTryReadFull(t *testing.T) {
	s := new(xstream.XStream).Init(100)

	ok, err := s.TryReadFull(0, func(data []byte) {
		assert.Len(t, data, 0)
	})
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = s.TryReadFull(101, func([]byte) {})
	assert.False(t, ok)
	assert.NoError(t, err)

	ok, err = s.TryReadFull(10, func([]byte) {})
	assert.False(t, ok)
	assert.NoError(t, err)

	ok, err = s.TryWriteAll(25, func(buffer []byte) {
		assert.Len(t, buffer, 25)
	})
	assert.True(t, ok)
	assert.NoError(t, err)

	for i := 0; i < 2; i++ {
		ok, err = s.TryReadFull(10, func(buffer []byte) {
			assert.Len(t, buffer, 10)
		})
		assert.True(t, ok)
		assert.NoError(t, err)
	}

	ok, err = s.TryReadFull(10, func([]byte) {})
	assert.False(t, ok)
	assert.NoError(t, err)

	ok, err = s.TryWriteAll(5, func(buffer []byte) {
		assert.Len(t, buffer, 5)
	})
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = s.TryReadFull(10, func(buffer []byte) {
		assert.Len(t, buffer, 10)
	})
	assert.True(t, ok)
	assert.NoError(t, err)

	err = s.Close()
	assert.NoError(t, err)

	ok, err = s.TryReadFull(10, func([]byte) {})
	assert.False(t, ok)
	assert.Error(t, err, xstream.ErrClosed)
}

func TestStreamWriteAll(t *testing.T) {
	s := new(xstream.XStream).Init(100)

	err := s.WriteAll(context.Background(), 0, func(buffer []byte) {
		assert.Len(t, buffer, 0)
	})
	assert.NoError(t, err)

	err = s.WriteAll(context.Background(), 101, func([]byte) {})
	assert.Error(t, err, xstream.ErrSizeExceeded)

	ok, err := s.TryWriteAll(100, func(buffer []byte) {
		assert.Len(t, buffer, 100)
	})
	assert.True(t, ok)
	assert.NoError(t, err)

	go func() {
		ok, err := s.TryReadFull(25, func(data []byte) {
			assert.Len(t, data, 25)
		})
		assert.True(t, ok)
		assert.NoError(t, err)
	}()

	for i := 0; i < 2; i++ {
		err = s.WriteAll(context.Background(), 10, func(buffer []byte) {
			assert.Len(t, buffer, 10)
		})
		assert.NoError(t, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err = s.WriteAll(ctx, 10, func(buffer []byte) {
		assert.Len(t, buffer, 10)
	})
	assert.Error(t, err, context.DeadlineExceeded)

	go func() {
		time.Sleep(50 * time.Millisecond)
		ok, err := s.TryReadFull(5, func(data []byte) {
			assert.Len(t, data, 5)
		})
		assert.True(t, ok)
		assert.NoError(t, err)
	}()

	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err = s.WriteAll(ctx, 10, func(buffer []byte) {
		assert.Len(t, buffer, 10)
	})
	assert.NoError(t, err)

	go func() {
		err := s.Close()
		assert.NoError(t, err)
	}()

	err = s.WriteAll(context.Background(), 10, func([]byte) {})
	assert.Error(t, err, xstream.ErrClosed)
	err = s.WriteAll(context.Background(), 0, func([]byte) {})
	assert.Error(t, err, xstream.ErrClosed)
}

func TestStreamTryWriteAll(t *testing.T) {
	s := new(xstream.XStream).Init(100)

	ok, err := s.TryWriteAll(0, func(buffer []byte) {
		assert.Len(t, buffer, 0)
	})
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = s.TryWriteAll(101, func([]byte) {})
	assert.False(t, ok)
	assert.NoError(t, err)

	ok, err = s.TryWriteAll(100, func(buffer []byte) {
		assert.Len(t, buffer, 100)
	})
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = s.TryWriteAll(10, func([]byte) {})
	assert.False(t, ok)
	assert.NoError(t, err)

	ok, err = s.TryReadFull(25, func(data []byte) {
		assert.Len(t, data, 25)
	})
	assert.True(t, ok)
	assert.NoError(t, err)

	for i := 0; i < 2; i++ {
		ok, err = s.TryWriteAll(10, func(buffer []byte) {
			assert.Len(t, buffer, 10)
		})
		assert.True(t, ok)
		assert.NoError(t, err)
	}

	ok, err = s.TryWriteAll(10, func([]byte) {})
	assert.False(t, ok)
	assert.NoError(t, err)

	ok, err = s.TryReadFull(5, func(data []byte) {
		assert.Len(t, data, 5)
	})
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = s.TryWriteAll(10, func(buffer []byte) {
		assert.Len(t, buffer, 10)
	})
	assert.True(t, ok)
	assert.NoError(t, err)

	err = s.Close()
	assert.NoError(t, err)

	ok, err = s.TryWriteAll(10, func([]byte) {})
	assert.False(t, ok)
	assert.Error(t, err, xstream.ErrClosed)
}

func TestStreamEnlarge(t *testing.T) {
	s := new(xstream.XStream).Init(100)
	assert.Equal(t, 100, s.Size())

	err := s.WriteAll(context.Background(), 101, func([]byte) {})
	assert.Error(t, err, xstream.ErrSizeExceeded)

	err = s.Enlarge(1)
	assert.NoError(t, err)
	assert.Equal(t, 101, s.Size())

	ok, err := s.TryWriteAll(101, func(buffer []byte) {
		assert.Len(t, buffer, 101)
	})
	assert.True(t, ok)
	assert.NoError(t, err)

	err = s.Close()
	assert.NoError(t, err)

	err = s.Enlarge(1)
	assert.Error(t, err, xstream.ErrClosed)
}

func TestStreamPeek(t *testing.T) {
	s := new(xstream.XStream).Init(100)

	err := s.Peek(func(data []byte) {
		assert.Len(t, data, 0)
	})
	assert.NoError(t, err)

	ok, err := s.TryWriteAll(100, func(buffer []byte) {
		assert.Len(t, buffer, 100)
		for i := range buffer {
			buffer[i] = uint8(i)
		}
	})
	assert.True(t, ok)
	assert.NoError(t, err)

	err = s.Peek(func(data []byte) {
		assert.Len(t, data, 100)
		f := true
		for i := range data {
			f = f && data[i] == uint8(i)
		}
		assert.True(t, f)
	})
	assert.NoError(t, err)
}

func TestStreamClose(t *testing.T) {
	s := new(xstream.XStream).Init(100)

	err := s.Close()
	assert.NoError(t, err)

	err = s.Close()
	assert.Error(t, err, xstream.ErrClosed)
}

func TestStreamWriteAndRead(t *testing.T) {
	s := new(xstream.XStream).Init(100)

	runWriter := func(wg *sync.WaitGroup, m int, n int, l int) {
		nWritten := int64(0)
		nDeadlineExceeded := int64(0)
		wg2 := sync.WaitGroup{}
		wg.Add(m + 1)
		wg2.Add(m)
		for i := 0; i < m; i++ {
			go func() {
				defer func() {
					wg.Done()
					wg2.Done()
				}()
				for i := 0; i < n; i++ {
					d := time.Now().Add(time.Duration(7-rand.Intn(10)) * time.Millisecond)
					ctx, cancel := context.WithDeadline(context.Background(), d)
					bs := 1 + rand.Intn(2*l)
					err := s.WriteAll(ctx, bs, func(buffer []byte) {
						for i := range buffer {
							buffer[i] = 111
						}
					})
					if err == nil {
						atomic.AddInt64(&nWritten, int64(bs))
					} else {
						if assert.Error(t, err, context.DeadlineExceeded) {
							atomic.AddInt64(&nDeadlineExceeded, 1)
						}
					}
					cancel()
				}
			}()
		}
		go func() {
			wg2.Wait()
			t.Logf("writer: nWritten=%v nDeadlineExceeded=%v", nWritten, nDeadlineExceeded)
			wg.Done()
		}()
	}

	runReaders := func(wg *sync.WaitGroup, m int, n int, l int) {
		nRead := int64(0)
		nDeadlineExceeded := int64(0)
		wg2 := sync.WaitGroup{}
		wg.Add(m + 1)
		wg2.Add(m)
		for i := 0; i < m; i++ {
			go func() {
				defer func() {
					wg.Done()
					wg2.Done()
				}()
				for i := 0; i < n; i++ {
					d := time.Now().Add(time.Duration(7-rand.Intn(10)) * time.Millisecond)
					ctx, cancel := context.WithDeadline(context.Background(), d)
					ds := 1 + rand.Intn(2*l)
					err := s.ReadFull(ctx, ds, func(data []byte) {
						f := true
						for i := range data {
							f = f && data[i] == 111
						}
						assert.True(t, f)
					})
					if err == nil {
						atomic.AddInt64(&nRead, int64(ds))
					} else {
						if assert.Error(t, err, context.DeadlineExceeded) {
							atomic.AddInt64(&nDeadlineExceeded, 1)
						}
					}
					cancel()
				}
			}()
		}
		go func() {
			wg2.Wait()
			t.Logf("reader: nRead=%v nDeadlineExceeded=%v", nRead, nDeadlineExceeded)
			wg.Done()
		}()
	}

	wg := sync.WaitGroup{}

	runWriter(&wg, 1, 100, 10)
	runReaders(&wg, 1, 100, 10)
	wg.Wait()

	runWriter(&wg, 10, 10, 10)
	runReaders(&wg, 10, 10, 10)
	wg.Wait()
}
