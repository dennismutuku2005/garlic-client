package garlic

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"garlic-client/internal/protocol"
)

// Expose internal protocol types to the public garlic package using aliases.
type Command = protocol.Command
type Reply = protocol.Reply
type TrapError = protocol.TrapError
type FatalError = protocol.FatalError

// Expose function aliases for convenience.
var NewCommand = protocol.NewCommand

// Option defines a functional configuration option for the RouterOS client.
type Option func(*Client)

// WithTLS configures the client to connect using secure TLS on the default API port (8729).
func WithTLS() Option {
	return func(c *Client) {
		c.tls = true
	}
}

// WithTimeout sets a custom connection timeout for TCP dial and login handshakes.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.timeout = d
	}
}

// Client is the main handler to communicate with a MikroTik router using RouterOS API.
type Client struct {
	address  string
	username string
	password string
	tls      bool
	timeout  time.Duration

	conn   net.Conn
	reader *bufio.Reader
	async  *protocol.AsyncManager
	ctx    context.Context
	cancel context.CancelFunc

	mu     sync.Mutex // Protects connection state and writing to socket
	closed bool
}

// New creates a new RouterOS client instance.
// If the port is omitted from the address, it defaults to 8728 (or 8729 if WithTLS is used).
func New(address, username, password string, opts ...Option) (*Client, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := &Client{
		address:  address,
		username: username,
		password: password,
		timeout:  10 * time.Second, // Default connection timeout
		ctx:      ctx,
		cancel:   cancel,
	}

	// Apply functional options
	for _, opt := range opts {
		opt(c)
	}

	// If no port is specified in the address, append the default port
	if _, _, err := net.SplitHostPort(c.address); err != nil {
		defaultPort := "8728"
		if c.tls {
			defaultPort = "8729"
		}
		c.address = net.JoinHostPort(c.address, defaultPort)
	}

	return c, nil
}

// Connect establishes the TCP/TLS connection to the router and completes the login handshake.
func (c *Client) Connect() error {
	c.mu.Lock()
	if c.conn != nil && !c.closed {
		c.mu.Unlock()
		return nil // already connected
	}
	c.closed = false
	c.mu.Unlock()

	var conn net.Conn
	var err error

	if c.tls {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		dialer := &net.Dialer{Timeout: c.timeout}
		conn, err = tls.DialWithDialer(dialer, "tcp", c.address, tlsConfig)
	} else {
		dialer := net.Dialer{Timeout: c.timeout}
		conn, err = dialer.DialContext(c.ctx, "tcp", c.address)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", c.address, err)
	}

	// 1. Perform login handshake synchronously using buffered reader and direct connection writer
	br := bufio.NewReader(conn)
	rw := struct {
		io.Reader
		io.Writer
	}{
		Reader: br,
		Writer: conn,
	}

	// Set deadline for the login handshake to prevent hanging on non-API ports (e.g. Winbox)
	conn.SetDeadline(time.Now().Add(c.timeout))

	if err := protocol.Login(rw, c.username, c.password); err != nil {
		conn.Close()
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return fmt.Errorf("login timed out")
		}
		return err
	}

	// Clear the deadline for normal background operation
	conn.SetDeadline(time.Time{})

	c.mu.Lock()
	c.conn = conn
	c.reader = br
	c.async = protocol.NewAsyncManager()
	c.mu.Unlock()

	// 2. Start the background listener loop
	go c.readLoop()

	return nil
}

// readLoop listens for reply sentences and routes them using the async manager.
// Recovers from unexpected panics to keep the parent application alive.
func (c *Client) readLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("garlic: recovered from unexpected panic in readLoop: %v", r)
		}
		c.Close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		sentence, err := protocol.ReadSentence(c.reader)
		if err != nil {
			// Socket read error or closed
			return
		}

		reply, err := protocol.ParseReply(sentence)
		if err != nil {
			// Ignore invalid sentence and try to keep reading
			continue
		}

		if reply.Type == "!fatal" {
			// Send fatal error to all active listeners and exit loop
			c.async.Distribute(reply)
			return
		}

		c.async.Distribute(reply)
	}
}

// Close gracefully closes the client connection and resources.
func (c *Client) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}

	var err error
	if c.conn != nil {
		err = c.conn.Close()
	}

	c.mu.Lock()
	c.conn = nil
	c.mu.Unlock()

	if c.async != nil {
		c.async.CloseAll()
	}

	return err
}

// Run executes a command synchronously and returns all the response `!re` replies.
func (c *Client) Run(cmd *Command) ([]*Reply, error) {
	if !c.IsConnected() {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	if cmd.Tag == "" {
		cmd.Tag = c.async.NextTag()
	}

	ch, _ := c.async.Register(cmd.Tag)
	defer c.async.Unregister(cmd.Tag)

	c.mu.Lock()
	if c.closed || c.conn == nil {
		c.mu.Unlock()
		return nil, protocol.ErrNotConnected
	}
	err := protocol.WriteSentence(c.conn, cmd.Sentence())
	c.mu.Unlock()

	if err != nil {
		return nil, fmt.Errorf("failed to send command sentence: %w", err)
	}

	var replies []*Reply
	for reply := range ch {
		if err := reply.AsError(); err != nil {
			return nil, err
		}
		if reply.Type == "!re" {
			replies = append(replies, reply)
		}
	}

	return replies, nil
}

// RunOne executes a command synchronously and returns the first `!re` reply.
func (c *Client) RunOne(cmd *Command) (*Reply, error) {
	replies, err := c.Run(cmd)
	if err != nil {
		return nil, err
	}
	if len(replies) == 0 {
		return nil, fmt.Errorf("no replies returned for command %s", cmd.Name)
	}
	return replies[0], nil
}

// RunAsync executes a command asynchronously and returns a channel delivering the replies.
// The listener is automatically closed once the command is complete.
func (c *Client) RunAsync(cmd *Command) (<-chan *Reply, error) {
	if !c.IsConnected() {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	if cmd.Tag == "" {
		cmd.Tag = c.async.NextTag()
	}

	ch, _ := c.async.Register(cmd.Tag)

	c.mu.Lock()
	if c.closed || c.conn == nil {
		c.mu.Unlock()
		c.async.Unregister(cmd.Tag)
		return nil, protocol.ErrNotConnected
	}
	err := protocol.WriteSentence(c.conn, cmd.Sentence())
	c.mu.Unlock()

	if err != nil {
		c.async.Unregister(cmd.Tag)
		return nil, fmt.Errorf("failed to send async command sentence: %w", err)
	}

	return ch, nil
}

// Cancel sends a cancellation request to the router for a running async command tag.
func (c *Client) Cancel(tag string) error {
	cancelCmd := NewCommand("/cancel").Arg("tag", tag)
	_, err := c.Run(cancelCmd)
	return err
}

// Ping tests the connection by running a command to fetch the system identity.
func (c *Client) Ping() error {
	if err := c.Connect(); err != nil {
		return err
	}
	_, err := c.RunOne(NewCommand("/system/identity/print"))
	return err
}

// IsConnected returns true if the connection is active and not closed.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil && !c.closed
}

// GetConnection returns the underlying net.Conn (useful for direct socket introspection/debugging).
func (c *Client) GetConnection() net.Conn {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn
}