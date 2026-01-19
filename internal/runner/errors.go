package runner

import "errors"

var (
	// ErrMaxTurnsExceeded is returned when the agent loop exceeds MaxTurns
	ErrMaxTurnsExceeded = errors.New("max turns exceeded")

	// ErrTimeout is returned when agent execution exceeds timeout
	ErrTimeout = errors.New("agent execution timeout")
)
