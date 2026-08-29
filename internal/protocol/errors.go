package protocol

import "fmt"

// TrapError is returned when the router replies with a !trap sentence.
// This typically indicates a command error (bad param, not found, etc.)
type TrapError struct {
	Status   string `json:"status"`
	Category string `json:"category,omitempty"`
	Message  string `json:"message"`
}

func (e *TrapError) Error() string {
	if e.Category != "" {
		return fmt.Sprintf("RouterOS trap: %s (category %s)", e.Message, e.Category)
	}
	return fmt.Sprintf("RouterOS trap: %s", e.Message)
}

// FatalError is returned when the router replies with a !fatal sentence.
// This usually means the connection will be closed by the router.
type FatalError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (e *FatalError) Error() string {
	return fmt.Sprintf("RouterOS fatal: %s", e.Message)
}

// ClientError wraps any generic or network error to format it as JSON.
type ClientError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *ClientError) Error() string {
	return e.Err.Error()
}

func (e *ClientError) Unwrap() error {
	return e.Err
}

// ErrNotConnected is returned when an operation is attempted on a disconnected client.
var ErrNotConnected = fmt.Errorf("garlic: client is not connected")
