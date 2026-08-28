package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/dennismutuku2005/garlic-client"
)

func main() {
	host := flag.String("host", "192.168.88.1", "MikroTik API host address")
	port := flag.Int("port", 0, "MikroTik API port (optional, 0 to use default)")
	user := flag.String("user", "admin", "MikroTik username")
	pass := flag.String("pass", "", "MikroTik password")
	tls := flag.Bool("tls", false, "Use TLS connection")
	flag.Parse()

	fmt.Printf("=== Connecting to RouterOS API at %s (port: %d) ===\n", *host, *port)

	var opts []garlic.Option
	if *tls {
		opts = append(opts, garlic.WithTLS())
	}
	if *port != 0 {
		opts = append(opts, garlic.WithPort(*port))
	}
	opts = append(opts, garlic.WithTimeout(5*time.Second))

	client, err := garlic.New(*host, *user, *pass, opts...)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect/authenticate: %v", err)
	}
	fmt.Println("Successfully connected and authenticated!")

	// Fetch system identity dynamically
	identReply, err := client.RunOne(garlic.NewCommand("/system/identity/print"))
	if err != nil {
		log.Printf("Failed to get system identity: %v", err)
	} else {
		fmt.Printf("Router Identity: %s\n", identReply.Map["name"])
	}

	// Fetch system resources dynamically
	resReply, err := client.RunOne(garlic.NewCommand("/system/resource/print"))
	if err != nil {
		log.Printf("Failed to get system resources: %v", err)
	} else {
		fmt.Printf("Router Version:  %s\n", resReply.Map["version"])
		fmt.Printf("Router CPU:      %s (%s cores)\n", resReply.Map["cpu"], resReply.Map["cpu-count"])
		fmt.Printf("CPU Load:        %s%%\n", resReply.Map["cpu-load"])
		fmt.Printf("Free Memory:     %s / %s bytes\n", resReply.Map["free-memory"], resReply.Map["total-memory"])
	}
}