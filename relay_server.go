package main

import (
	// "context"
	"fmt"
	"log"
	"os"

	libp2p "github.com/libp2p/go-libp2p"
	circuitv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
)

func main() {
	// ctx := context.Background()

	port := os.Getenv("PORT") // Render maps this to 443 automatically
	if port == "" {
		port = "443"
	}

	hostname := os.Getenv("RENDER_EXTERNAL_HOSTNAME")
	if hostname == "" {
		hostname = "localhost"
	}

	h, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s/wss", port)),
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = circuitv2.New(h)
	if err != nil {
		log.Fatal(err)
	}

	publicAddr := fmt.Sprintf("/ip4/%s/tcp/%s/wss/p2p/%s", hostname, port, h.ID())

	fmt.Println("✅ Relay server started!")
	fmt.Println("PeerID:", h.ID())
	fmt.Println("Public Multiaddr (give this to clients):")
	fmt.Println(publicAddr)

	select {} // keep running
}
