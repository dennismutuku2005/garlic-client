package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/dennismutuku2005/garlic-client"
)

// PPPoESecret represents a PPP secret in RouterOS.
type PPPoESecret struct {
	ID            string `json:".id,omitempty"`
	Name          string `json:"name"`
	Service       string `json:"service"`
	Profile       string `json:"profile"`
	LocalAddress  string `json:"local-address,omitempty"`
	RemoteAddress string `json:"remote-address,omitempty"`
	LastCaller    string `json:"last-caller,omitempty"`
	Comment       string `json:"comment,omitempty"`
	Disabled      string `json:"disabled"`
}

func main() {
	host := flag.String("host", "167.99.9.95", "MikroTik API host address")
	port := flag.Int("port", 0, "MikroTik API port (optional, 0 to use default)")
	user := flag.String("user", "apiuser", "MikroTik username")
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

	fmt.Println("\n1. Fetching all PPPoE secrets...")
	
	// Query /ppp/secret/print and filter by service=pppoe
	cmd := garlic.NewCommand("/ppp/secret/print").Filter("service", "pppoe")
	
	replies, err := client.Run(cmd)
	if err != nil {
		log.Fatalf("Failed to fetch PPPoE secrets: %v", err)
	}

	fmt.Printf("Received %d replies from the router.\n", len(replies))

	secrets := make([]PPPoESecret, 0, len(replies))
	for _, r := range replies {
		s := PPPoESecret{
			ID:            r.Map[".id"],
			Name:          r.Map["name"],
			Service:       r.Map["service"],
			Profile:       r.Map["profile"],
			LocalAddress:  r.Map["local-address"],
			RemoteAddress: r.Map["remote-address"],
			LastCaller:    r.Map["last-caller"],
			Comment:       r.Map["comment"],
			Disabled:      r.Map["disabled"],
		}
		secrets = append(secrets, s)
	}

	// 2. Print raw map structure
	fmt.Println("\n=== Raw Reply Map Format (Key-Value) ===")
	if len(replies) == 0 {
		fmt.Println("  (No PPPoE secrets found on the router)")
	} else {
		for i, r := range replies {
			fmt.Printf("Reply %d Map:\n", i+1)
			for k, v := range r.Map {
				fmt.Printf("  %s: %s\n", k, v)
			}
			fmt.Println()
		}
	}

	// 3. Print clean serialized JSON format
	fmt.Println("=== Clean Serialized JSON Format ===")
	jsonData, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		log.Fatalf("Failed to serialize secrets to JSON: %v", err)
	}
	fmt.Println(string(jsonData))
}
