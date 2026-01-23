package stream

import (
	"sync"
	"testing"
)

func TestSequenceNumber_Next(t *testing.T) {
	seq := NewSequenceNumber()

	// Test sequential calls
	for i := 0; i < 10; i++ {
		if got := seq.Next(); got != i {
			t.Errorf("Next() = %v, want %v", got, i)
		}
	}
}

func TestSequenceNumber_Current(t *testing.T) {
	seq := NewSequenceNumber()

	// Initial value should be 0
	if got := seq.Current(); got != 0 {
		t.Errorf("Current() = %v, want %v", got, 0)
	}

	// After calling Next(), Current should reflect the last returned value + 1
	seq.Next() // 0
	seq.Next() // 1
	if got := seq.Current(); got != 2 {
		t.Errorf("Current() = %v, want %v", got, 2)
	}
}

func TestSequenceNumber_Concurrent(t *testing.T) {
	seq := NewSequenceNumber()
	const goroutines = 100
	const iterations = 100

	var wg sync.WaitGroup
	results := make([]int, goroutines*iterations)
	resultsMu := sync.Mutex{}
	resultIdx := 0

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				num := seq.Next()
				resultsMu.Lock()
				results[resultIdx] = num
				resultIdx++
				resultsMu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Verify all numbers are unique
	seen := make(map[int]bool)
	for _, num := range results {
		if seen[num] {
			t.Errorf("Duplicate sequence number: %d", num)
		}
		seen[num] = true
	}

	// Verify we got all numbers from 0 to (goroutines*iterations - 1)
	if len(seen) != goroutines*iterations {
		t.Errorf("Expected %d unique numbers, got %d", goroutines*iterations, len(seen))
	}
}
