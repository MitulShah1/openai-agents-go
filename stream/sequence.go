package stream

import "sync/atomic"

// SequenceNumber provides thread-safe sequence number generation for events.
// This ensures events can be properly ordered even when generated concurrently.
type SequenceNumber struct {
	counter atomic.Int64
}

// NewSequenceNumber creates a new sequence number generator starting at 0.
func NewSequenceNumber() *SequenceNumber {
	return &SequenceNumber{}
}

// Next returns the next sequence number and increments the counter.
// This method is thread-safe and can be called concurrently.
func (s *SequenceNumber) Next() int {
	return int(s.counter.Add(1) - 1)
}

// Current returns the current sequence number without incrementing.
func (s *SequenceNumber) Current() int {
	return int(s.counter.Load())
}
