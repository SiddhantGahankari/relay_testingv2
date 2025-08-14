package main

import (
	// "context"
	"fmt"
	"log"
	"net"
	"os"

	libp2p "github.com/libp2p/go-libp2p"
	circuitv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
)

func main() {
	// ctx := context.Background()

	port := os.Getenv("PORT")
	if port == "" {
		port = "9001" // fallback for local
	}
	hostname := os.Getenv("RENDER_EXTERNAL_HOSTNAME")
	if hostname == "" {
		hostname = "localhost"
	}

	// Resolve hostname to IPv4
	ipAddrs, err := net.LookupIP(hostname)
	if err != nil || len(ipAddrs) == 0 {
		log.Fatalf("Could not resolve public IP for %s: %v", hostname, err)
	}
	publicIP := ipAddrs[0].String()

	h, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s/ws", port)),
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = circuitv2.New(h)
	if err != nil {
		log.Fatal(err)
	}

	publicAddr := fmt.Sprintf("/ip4/%s/tcp/%s/ws/p2p/%s", publicIP, port, h.ID())

	fmt.Println("✅ Relay server started!")
	fmt.Println("PeerID:", h.ID())
	fmt.Println("Public Multiaddr (give this to clients):")
	fmt.Println(publicAddr)

	select {}
}
