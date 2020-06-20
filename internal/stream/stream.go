package stream

import "fmt"

type Stream struct {
	base        []byte
	dataIndex   int
	bufferIndex int
}

func (s *Stream) AcquireBuffer(bufferSize int) []byte {
	if buffer := s.buffer(); bufferSize <= len(buffer) {
		return buffer[:bufferSize]
	}

	data := s.Data()

	if bufferSize <= len(s.base)-len(data) {
		if bufferSize >= len(data) {
			s.setData(data)
			return s.buffer()[:bufferSize]
		}

		s.base = make([]byte, 2*len(s.base))
		s.setData(data)
		return s.buffer()[:bufferSize]
	}

	s.base = make([]byte, nextPowerOfTwo64(int64(len(data)+bufferSize)))
	s.setData(data)
	return s.buffer()[:bufferSize]
}

func (s *Stream) CommitData(dataSize int) {
	if dataSize > s.bufferSize() {
		panic(fmt.Errorf("stream: illegal argument; dataSize=%d", dataSize))
	}

	s.bufferIndex += dataSize
}

func (s *Stream) DiscardData(dataSize int) {
	if dataSize > s.DataSize() {
		panic(fmt.Errorf("stream: illegal argument; dataSize=%d", dataSize))
	}

	s.dataIndex += dataSize

	if data := s.Data(); len(data) <= s.dataIndex {
		s.setData(data)
	}
}

func (s *Stream) Data() []byte {
	return s.base[s.dataIndex:s.bufferIndex]
}

func (s *Stream) DataSize() int {
	return s.bufferIndex - s.dataIndex
}

func (s *Stream) setData(data []byte) {
	s.dataIndex = 0
	s.bufferIndex = copy(s.base, data)
}

func (s *Stream) buffer() []byte {
	return s.base[s.bufferIndex:]
}

func (s *Stream) bufferSize() int {
	return len(s.base) - s.bufferIndex
}

func nextPowerOfTwo64(x int64) int64 {
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x |= x >> 32
	x++
	return x
}
