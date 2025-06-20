package common

import (
	"bytes"
	"sync"
)

type CircularBuffer struct {
	mu      sync.Mutex
	buffers [][]byte
	size    int
	index   int
	full    bool
}

// NewCircularBuffer creates a new (rolling) CircularBuffer with the specified capacity.
//
// Example usage:
// cb := common.NewCircularBuffer(100)
func NewCircularBuffer(capacity int) *CircularBuffer {
	return &CircularBuffer{
		buffers: make([][]byte, capacity),
		size:    capacity,
	}
}

// Write adds data to the CircularBuffer, overwriting the oldest data if the buffer is full.
// The data is copied to ensure that the original data can be modified without affecting the buffer.
//
// Example usage: cb.Write([]byte("some data"))
func (cb *CircularBuffer) Write(data []byte) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.buffers[cb.index] = append([]byte(nil), data...)
	cb.index = (cb.index + 1) % cb.size
	if cb.index == 0 {
		cb.full = true
	}
}

// Read retrieves the data from the CircularBuffer.
//
// Example usage: data := cb.Bytes()
func (cb *CircularBuffer) Bytes() []byte {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	var out bytes.Buffer
	if cb.full {
		for i := cb.index; i < cb.size; i++ {
			out.Write(cb.buffers[i])
		}
	}
	for i := 0; i < cb.index; i++ {
		out.Write(cb.buffers[i])
	}
	return out.Bytes()
}
