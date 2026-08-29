/*
Package garlic is a clean, robust, concurrency-safe, and highly optimized Go client for the MikroTik RouterOS API.

It implements the RouterOS API protocol over both TCP and TLS, featuring safe connection handling,
an asynchronous multiplexer for running concurrent queries over a single connection, and support
for structured error parsing (including trap, fatal, and connection/client errors).

# Installation

To import this package into your Go project, run:

	go get github.com/dennismutuku2005/garlic-client

Then import it in your Go source files:

	import "github.com/dennismutuku2005/garlic-client"

# Quick Start

Here is a basic example of initializing a client, connecting to a router, and running a simple command:

	package main

	import (
		"fmt"
		"log"
		"time"

		"github.com/dennismutuku2005/garlic-client"
	)

	func main() {
		// Initialize client using functional configuration options
		client, err := garlic.New(
			"192.168.88.1",
			"admin",
			"password",
			garlic.WithTimeout(5*time.Second), // Custom connection/handshake timeout
		)
		if err != nil {
			log.Fatalf("Initialization failed: %v", err)
		}
		defer client.Close()

		// Establish connection and perform RouterOS login handshake
		if err := client.Connect(); err != nil {
			log.Fatalf("Connection failed: %v", err)
		}

		// Fetch system identity
		reply, err := client.RunOne(garlic.NewCommand("/system/identity/print"))
		if err != nil {
			log.Fatalf("Command execution failed: %v", err)
		}

		fmt.Printf("Router Identity: %s\n", reply.Map["name"])
	}

# Execution Modes

The Client provides three main execution modes:

1. Synchronous List Execution: Run()
Executes a command and blocks until all replies (!re) are collected and the !done state is reached.

	cmd := garlic.NewCommand("/ip/address/print")
	replies, err := client.Run(cmd)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	for _, r := range replies {
		fmt.Printf("Interface: %s, Address: %s\n", r.Map["interface"], r.Map["address"])
	}

2. Single-Row Execution: RunOne()
Executes a command and returns only the first reply sentence (!re). Perfect for commands yielding a single item.

	reply, err := client.RunOne(garlic.NewCommand("/system/resource/print"))
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("Board Name: %s, Uptime: %s\n", reply.Map["board-name"], reply.Map["uptime"])

3. Asynchronous Execution: RunAsync()
Runs a command concurrently and returns a channel immediately. Replies are delivered on the channel as they arrive.
Ideal for streaming or long-running commands (e.g. /interface/monitor-traffic).

	cmd := garlic.NewCommand("/interface/monitor-traffic").Arg("interface", "ether1")
	ch, err := client.RunAsync(cmd)
	if err != nil {
		log.Fatalf("Error starting async command: %v", err)
	}

	for reply := range ch {
		// Always check for sentence level errors
		if err := reply.AsError(); err != nil {
			log.Printf("Sentence error: %v", err)
			break
		}
		fmt.Printf("Rx: %s bps, Tx: %s bps\n", reply.Map["rx-bits-per-second"], reply.Map["tx-bits-per-second"])
	}

# Command Builder API

Build parameters and options fluently using chainable methods on a Command:

  - Arg(key, value string): Adds general arguments to the command (=name=value).
  - Filter(key, value string): Adds search filters (?name=value).
  - Proplist(fields ...string): Requests specific fields (.proplist=field1,field2).
  - Attr(value string): Appends a raw attribute or flag.

Example:

	cmd := garlic.NewCommand("/ip/route/print").
		Filter("active", "true").
		Proplist("dst-address", "gateway")

# Structured Error Handling

All client and router errors are wrapped or classified into specific types so they can be handled programmatically:

1. TrapError: Returned when RouterOS rejects a command (e.g. invalid arguments or insufficient privileges). Contains Status, Category and Message.
2. FatalError: Returned when the router experiences a severe fault or closes the connection suddenly.
3. ClientError: Wraps underlying networking, DNS, serialization, or connection issues. Supports standard JSON serialization.

You can inspect errors using errors.As:

	replies, err := client.Run(cmd)
	if err != nil {
		var trapErr *garlic.TrapError
		if errors.As(err, &trapErr) {
			fmt.Printf("Router Trap: %s (Category %s)\n", trapErr.Message, trapErr.Category)
			return
		}
		log.Fatalf("Other error: %v", err)
	}

# Client Options

When initializing a new client with garlic.New, you can supply configuration options:

	client, err := garlic.New("router.local", "admin", "password",
		garlic.WithTLS(),                   // Use encrypted TLS API port (defaults to 8729)
		garlic.WithPort(9000),              // Use a custom non-standard port
		garlic.WithTimeout(5*time.Second),  // Set connection and login handshake timeout
	)
*/
package garlic
