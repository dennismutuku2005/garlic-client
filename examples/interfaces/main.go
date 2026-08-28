package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"garlic-client"
)

func main() {
	host := flag.String("host", "192.168.88.1", "MikroTik API host address")
	port := flag.Int("port", 0, "MikroTik API port (optional, 0 to use default)")
	user := flag.String("user", "admin", "MikroTik username")
	pass := flag.String("pass", "", "MikroTik password")
	tls := flag.Bool("tls", false, "Use TLS connection")
	flag.Parse()

	fmt.Printf("=== Fetching Interfaces dynamically from %s (port: %d) ===\n", *host, *port)

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

	// Query /interface/print dynamically
	cmd := garlic.NewCommand("/interface/print")
	replies, err := client.Run(cmd)
	if err != nil {
		log.Fatalf("Failed to retrieve interface list: %v", err)
	}

	fmt.Printf("\n%-5s %-20s %-10s %-8s %-8s %-8s\n", "ID", "Name", "Type", "MTU", "Running", "Disabled")
	fmt.Println("----------------------------------------------------------------------")
	for _, reply := range replies {
		fmt.Printf("%-5s %-20s %-10s %-8s %-8s %-8s\n",
			reply.Map[".id"],
			reply.Map["name"],
			reply.Map["type"],
			reply.Map["mtu"],
			reply.Map["running"],
			reply.Map["disabled"],
		)
	}
}
