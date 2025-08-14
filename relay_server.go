// relay_server.go
package main

import (
	"fmt"
	"log"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	circuitv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
)

func main() {
	// ctx := context.Background()

	// Create relay host listening on 0.0.0.0:PORT
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/9000",
			"/ip4/0.0.0.0/tcp/9001/ws",
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Start Circuit Relay v2 service
	_, err = circuitv2.New(h)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Relay server started!")
	fmt.Println("PeerID:", h.ID())
	for _, addr := range h.Addrs() {
		fmt.Println("Listening on:", addr)
	}

	// Keep running
	select {}
}

// Helper to print full multiaddrs
func fullAddrs(h host.Host) []string {
	var out []string
	for _, addr := range h.Addrs() {
		out = append(out, fmt.Sprintf("%s/p2p/%s", addr, h.ID()))
	}
	return out
}
