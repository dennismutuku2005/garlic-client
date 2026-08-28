package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"garlic-client"
)

func main() {
	addr := flag.String("addr", "167.99.9.95:9001", "MikroTik API address (ip:port)")
	user := flag.String("user", "admin", "MikroTik username")
	pass := flag.String("pass", "", "MikroTik password")
	tls := flag.Bool("tls", false, "Use TLS connection")
	flag.Parse()

	fmt.Printf("=== Connecting to RouterOS API at %s ===\n", *addr)
	config := garlic.Config{
		Address:  *addr,
		Username: *user,
		Password: *pass,
		TLS:      *tls,
		Timeout:  5 * time.Second,
	}

	client, err := garlic.New(config)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}
	defer client.Close()

	if err := client.Connect(); err != nil {
		log.Fatalf("❌ Failed to connect/authenticate: %v", err)
	}
	fmt.Println("✅ Successfully connected and authenticated!")

	// Fetch system identity
	ident, err := client.SystemIdentity()
	if err != nil {
		log.Printf("⚠️ Failed to get system identity: %v", err)
	} else {
		fmt.Printf("Router Identity: %s\n", ident.Name)
	}

	// Fetch system resources
	res, err := client.SystemResource()
	if err != nil {
		log.Printf("⚠️ Failed to get system resources: %v", err)
	} else {
		fmt.Printf("Router Version:  %s\n", res.Version)
		fmt.Printf("Router CPU:      %s (%s cores)\n", res.CPU, res.CPUCount)
		fmt.Printf("CPU Load:        %s%%\n", res.CPULoad)
		fmt.Printf("Free Memory:     %s / %s bytes\n", res.FreeMemory, res.TotalMemory)
	}
}