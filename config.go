package garlic

import (
	"fmt"
	"time"
)

// Config holds all connection settings for a MikroTik router.
type Config struct {
	Address  string        // IP:Port, e.g., "192.168.88.1:8728"
	Username string
	Password string
	TLS      bool          // true for port 8729 (TLS), false for 8728 (plain TCP)
	Timeout  time.Duration // Connection/handshake timeout
}

// Validate checks if the config contains all mandatory values.
func (cfg *Config) Validate() error {
	if cfg.Address == "" {
		return fmt.Errorf("garlic: address is required")
	}
	if cfg.Username == "" {
		return fmt.Errorf("garlic: username is required")
	}
	return nil
}
