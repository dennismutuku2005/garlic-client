package protocol

import "fmt"

// TrapError is returned when the router replies with a !trap sentence.
// This typically indicates a command error (bad param, not found, etc.)
type TrapError struct {
	Category string // numeric category code from the router
	Message  string // human-readable error message
}

func (e *TrapError) Error() string {
	return e.Message
}

// FatalError is returned when the router replies with a !fatal sentence.
// This usually means the connection will be closed by the router.
type FatalError struct {
	Message string
}

func (e *FatalError) Error() string {
	return e.Message
}

// ErrNotConnected is returned when an operation is attempted on a disconnected client.
var ErrNotConnected = fmt.Errorf("garlic: client is not connected")
