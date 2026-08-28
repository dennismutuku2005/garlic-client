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
	
	bridgeName := flag.String("bridge", "bridge_test", "Name of the new test bridge interface")
	poolName := flag.String("pool-name", "dhcp_pool_new", "DHCP address pool name")
	poolRange := flag.String("pool-range", "192.168.98.10-192.168.98.90", "IP address pool range")
	subnet := flag.String("subnet", "192.168.98.0/24", "Subnet for the DHCP server network")
	gateway := flag.String("gateway", "192.168.98.1", "Gateway IP for the subnet")
	dns := flag.String("dns", "8.8.8.8,1.1.1.1", "DNS servers comma-separated")
	
	flag.Parse()

	fmt.Printf("=== Creating Clean DHCP Server Configuration on %s ===\n", *host)

	// Create client
	client, err := garlic.New(*host, *user, *pass, garlic.WithPort(*port), garlic.WithTimeout(5*time.Second))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect/authenticate: %v", err)
	}
	fmt.Println("Successfully connected and authenticated!")

	// 1. Create a new test bridge: /interface/bridge/add
	fmt.Printf("1. Creating test bridge interface %q...\n", *bridgeName)
	addBridgeCmd := garlic.NewCommand("/interface/bridge/add").Arg("name", *bridgeName)
	_, err = client.Run(addBridgeCmd)
	if err != nil {
		log.Printf("   Warning/Error creating bridge (might already exist): %v", err)
	} else {
		fmt.Println("   Bridge interface created successfully!")
	}

	// 2. Create the IP Address Pool: /ip/pool/add
	fmt.Printf("2. Adding IP Pool %q (%s)...\n", *poolName, *poolRange)
	addPoolCmd := garlic.NewCommand("/ip/pool/add").
		Arg("name", *poolName).
		Arg("ranges", *poolRange)
	
	_, err = client.Run(addPoolCmd)
	if err != nil {
		log.Printf("   Warning/Error adding IP pool: %v", err)
	} else {
		fmt.Println("   IP Pool added successfully!")
	}

	// 3. Create the DHCP Network Configuration: /ip/dhcp-server/network/add
	fmt.Printf("3. Adding DHCP Network %q (gateway: %s)...\n", *subnet, *gateway)
	addNetCmd := garlic.NewCommand("/ip/dhcp-server/network/add").
		Arg("address", *subnet).
		Arg("gateway", *gateway).
		Arg("dns-server", *dns)
	
	_, err = client.Run(addNetCmd)
	if err != nil {
		log.Printf("   Warning/Error adding DHCP network: %v", err)
	} else {
		fmt.Println("   DHCP Network configuration added successfully!")
	}

	// 4. Create the DHCP Server: /ip/dhcp-server/add
	serverName := "dhcp_server_new"
	fmt.Printf("4. Adding DHCP Server %q on interface %q...\n", serverName, *bridgeName)
	addDhcpCmd := garlic.NewCommand("/ip/dhcp-server/add").
		Arg("name", serverName).
		Arg("interface", *bridgeName).
		Arg("address-pool", *poolName).
		Arg("disabled", "no")
	
	_, err = client.Run(addDhcpCmd)
	if err != nil {
		log.Fatalf("Fatal: Failed to add DHCP Server: %v", err)
	}
	fmt.Printf("✅ DHCP Server %q has been successfully created on %s!\n", serverName, *bridgeName)
}
