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

	fmt.Printf("=== Fetching Interfaces from %s ===\n", *addr)

	var opts []garlic.Option
	if *tls {
		opts = append(opts, garlic.WithTLS())
	}
	opts = append(opts, garlic.WithTimeout(5*time.Second))

	client, err := garlic.New(*addr, *user, *pass, opts...)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect/authenticate: %v", err)
	}

	ifaces, err := client.InterfaceList()
	if err != nil {
		log.Fatalf("Failed to retrieve interface list: %v", err)
	}

	fmt.Printf("\n%-5s %-20s %-10s %-8s %-8s %-8s\n", "ID", "Name", "Type", "MTU", "Running", "Disabled")
	fmt.Println("----------------------------------------------------------------------")
	for _, iface := range ifaces {
		fmt.Printf("%-5s %-20s %-10s %-8s %-8t %-8t\n",
			iface.ID,
			iface.Name,
			iface.Type,
			iface.MTU,
			iface.Running,
			iface.Disabled,
		)
	}
}
