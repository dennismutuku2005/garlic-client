<p align="center">
  <img src="../assets/logo.png" alt="garlic-client logo" width="300" />
</p>

<h1 align="center">garlic-client — Developer Documentation</h1>

<p align="center">
  Detailed guide, architecture, and API usage references for the Go RouterOS API Client.
</p>

---

## Architecture Overview

`garlic-client` is designed to be highly concurrency-safe, performant, and memory-efficient. Instead of blocking the client thread during network operations, it uses a background multiplexer goroutine.

### 1. Connection Flow
```mermaid
sequenceDiagram
    participant App as Go Application
    participant Client as garlic.Client
    participant Router as MikroTik RouterOS
    
    App->>Client: garlic.New(host, user, pass, options...)
    App->>Client: Client.Connect()
    Note over Client,Router: Synchronous Handshake Phase
    Client->>Router: TCP/TLS Dial Connection
    Client->>Router: Send /login name=user password=pass
    Router-->>Client: Reply !done (with MD5 challenge if legacy RouterOS v6)
    Note over Client,Router: Handshake finishes, background parser starts
    Client->>Client: Spawn Background readLoop()
    Client-->>App: Return success (nil error)
```

---

## API Reference

### 1. Initialization
Create a client instance using functional options:
```go
client, err := garlic.New("167.99.9.95", "admin", "password", 
	garlic.WithPort(9001), 
	garlic.WithTimeout(5*time.Second),
)
```

#### Available Options:
- **`WithTLS()`**: Connect securely via TLS (defaults to port `8729`).
- **`WithPort(port int)`**: Override host address port with a custom port.
- **`WithTimeout(d time.Duration)`**: Custom TCP/Handshake timeout.

---

### 2. Command Runner API

#### Dynamic Command Builder
Create commands using a fluent API styled like raw RouterOS commands:
```go
cmd := garlic.NewCommand("/ip/address/print").
	Filter("disabled", "false").
	Proplist("address", "interface")
```

#### Synchronous Execution: `Run()`
Executes the command and blocks until all `!re` replies are returned:
```go
replies, err := client.Run(cmd)
for _, reply := range replies {
	fmt.Println("IP:", reply.Map["address"])
}
```

#### Single Row Execution: `RunOne()`
Useful for commands where you only expect one response item:
```go
reply, err := client.RunOne(garlic.NewCommand("/system/identity/print"))
fmt.Println("Identity:", reply.Map["name"])
```

#### Asynchronous Execution: `RunAsync()`
Runs a command concurrently and returns a channel to read replies on-the-fly:
```go
ch, err := client.RunAsync(cmd)
for reply := range ch {
	if err := reply.AsError(); err != nil {
		log.Println("Error:", err)
		break
	}
	log.Println("Received item:", reply.Map["name"])
}
```

---

## Concurrency Safety & Multiplexing

The client supports running **multiple queries simultaneously** over a single TCP socket. Each query registers a unique `.tag` via `asyncManager`:

```mermaid
graph TD
    A[Goroutine 1: Run cmd1] -- Send tag1 --> C[Socket Writer]
    B[Goroutine 2: Run cmd2] -- Send tag2 --> C
    C --> D[MikroTik Router]
    D -- Reply tag2 --> E[readLoop Reader]
    D -- Reply tag1 --> E
    E --> F[AsyncManager Tag Router]
    F -- Route tag1 --> G[Goroutine 1 Channel]
    F -- Route tag2 --> H[Goroutine 2 Channel]
```

### Safety Features:
- **Panic Protection**: The background reader loop recovers from panics if a corrupted sentence is read, safely closing the connection instead of crashing the process.
- **Non-blocking Dispatch**: If a goroutine is slow to read from an async channel, the tag router automatically drops the blocked listener to prevent stalling the entire socket multiplexer.
