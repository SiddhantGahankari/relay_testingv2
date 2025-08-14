package main

import (
	"fmt"
	"log"
	"os"

	libp2p "github.com/libp2p/go-libp2p"
	circuitv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "9001" // Render detected ws port
	}
	hostname := os.Getenv("RENDER_EXTERNAL_HOSTNAME")
	if hostname == "" {
		hostname = "localhost"
	}

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

	publicAddr := fmt.Sprintf("/dns4/%s/tcp/%s/ws/p2p/%s", hostname, port, h.ID())

	fmt.Println("✅ Relay server started!")
	fmt.Println("PeerID:", h.ID())
	fmt.Println("Public Multiaddr (give this to clients):")
	fmt.Println(publicAddr)

	select {}
}
