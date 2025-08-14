package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	relay "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	ma "github.com/multiformats/go-multiaddr"
)

type RelayDist struct {
	relayID string
	dist    *big.Int
}

const ChatProtocol = protocol.ID("/chat/1.0.0")

type reqFormat struct {
	Type      string          `json:"type,omitempty"`
	PeerID    string          `json:"peer_id"`
	ReqParams json.RawMessage `json:"reqparams,omitempty"`
	Body      json.RawMessage `json:"body,omitempty"`
}

type respFormat struct {
	Type string `json:"type"`
	Resp []byte `json:"resp"`
}

var (
	ConnectedPeers   []string
	mu               sync.RWMutex
	RelayHost        host.Host
	OwnRelayAddrFull string

	// In-memory relay address list
	RelayMultiAddrList []string
)

type RelayEvents struct{}

func (re *RelayEvents) Listen(net network.Network, addr ma.Multiaddr)      {}
func (re *RelayEvents) ListenClose(net network.Network, addr ma.Multiaddr) {}
func (re *RelayEvents) Connected(net network.Network, conn network.Conn) {
	fmt.Printf("[INFO] Peer connected: %s\n", conn.RemotePeer())
}
func (re *RelayEvents) Disconnected(net network.Network, conn network.Conn) {
	fmt.Printf("[INFO] Peer disconnected: %s\n", conn.RemotePeer())
	mu.Lock()
	if contains(ConnectedPeers, conn.RemotePeer().String()) {
		remove(&ConnectedPeers, conn.RemotePeer().String())
	}
	mu.Unlock()
}

func main() {
	fmt.Println("STARTING RELAY CODE")

	// Hardcoded in-memory relay addresses (if you want to add others)
	RelayMultiAddrList = []string{ /* "/dns4/example.com/tcp/443/wss/p2p/<peerID>" */ }

	fmt.Println("[DEBUG] Creating connection manager...")
	connMgr, err := connmgr.NewConnManager(100, 400)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create connection manager: %v", err)
	}

	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		panic(err)
	}

	fmt.Println("[DEBUG] Creating relay host...")
	RelayHost, err = libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/443/ws"),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.ConnectionManager(connMgr),
		libp2p.EnableNATService(),
		libp2p.EnableRelayService(),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(websocket.New),
	)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create relay host: %v", err)
	}
	RelayHost.Network().Notify(&RelayEvents{})

	OwnRelayAddrFull = fmt.Sprintf("/dns4/libr-relay.onrender.com/tcp/443/wss/p2p/%s", RelayHost.ID().String())

	customRelayResources := relay.Resources{
		Limit: &relay.RelayLimit{
			Duration: 30 * time.Minute,
			Data:     1 << 20,
		},
		ReservationTTL:         time.Hour,
		MaxReservations:        1000,
		MaxCircuits:            64,
		BufferSize:             4096,
		MaxReservationsPerPeer: 10,
		MaxReservationsPerIP:   400,
		MaxReservationsPerASN:  64,
	}

	fmt.Println("[DEBUG] Enabling circuit relay service...")
	_, err = relay.New(RelayHost, relay.WithResources(customRelayResources))
	if err != nil {
		log.Fatalf("[ERROR] Failed to enable relay service: %v", err)
	}

	fmt.Printf("[INFO] Relay started!\n")
	fmt.Printf("[INFO] Peer ID: %s\n", RelayHost.ID())
	for _, addr := range RelayHost.Addrs() {
		fmt.Printf("[INFO] Relay Address: %s/p2p/%s\n", addr, RelayHost.ID())
	}

	RelayHost.SetStreamHandler(ChatProtocol, handleChatStream)
	go func() {
		for {
			fmt.Println(ConnectedPeers)
			time.Sleep(30 * time.Second)
		}
	}()

	fmt.Println("[DEBUG] Waiting for interrupt signal...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("[INFO] Shutting down relay...")
}

// ----------------- Chat Stream Handler -----------------
func handleChatStream(s network.Stream) {
	fmt.Println("[DEBUG] New chat stream opened")
	defer s.Close()
	// You can add your chat handling logic here
}

// ----------------- Utility Functions -----------------
func remove(list *[]string, val string) {
	for i, item := range *list {
		if item == val {
			*list = append((*list)[:i], (*list)[i+1:]...)
			return
		}
	}
}

func contains(arr []string, target string) bool {
	for _, vals := range arr {
		if vals == target {
			return true
		}
	}
	return false
}

func XorHexToBigInt(hex1, hex2 string) *big.Int {
	bytes1, err1 := hex.DecodeString(hex1)
	bytes2, err2 := hex.DecodeString(hex2)
	if err1 != nil || err2 != nil {
		log.Fatalf("Error decoding hex: %v %v", err1, err2)
	}
	if len(bytes1) != len(bytes2) {
		log.Fatalf("Hex strings must be the same length")
	}
	xorBytes := make([]byte, len(bytes1))
	for i := 0; i < len(bytes1); i++ {
		xorBytes[i] = bytes1[i] ^ bytes2[i]
	}
	return new(big.Int).SetBytes(xorBytes)
}

func GetRelayAddr(peerID string) string {
	if len(RelayMultiAddrList) == 0 {
		return OwnRelayAddrFull
	}
	var relayList []string
	for _, multiaddr := range RelayMultiAddrList {
		parts := strings.Split(multiaddr, "/")
		relayList = append(relayList, parts[len(parts)-1])
	}

	var distmap []RelayDist
	h1 := sha256.New()
	h1.Write([]byte(peerID))
	peerIDhash := hex.EncodeToString(h1.Sum(nil))

	for _, relay := range relayList {
		hR := sha256.New()
		hR.Write([]byte(relay))
		RelayIDhash := hex.EncodeToString(hR.Sum(nil))
		dist := XorHexToBigInt(peerIDhash, RelayIDhash)
		distmap = append(distmap, RelayDist{dist: dist, relayID: relay})
	}

	sort.Slice(distmap, func(i, j int) bool {
		return distmap[i].dist.Cmp(distmap[j].dist) < 0
	})

	relayIDused := distmap[0].relayID
	for _, multiaddr := range RelayMultiAddrList {
		parts := strings.Split(multiaddr, "/")
		if parts[len(parts)-1] == relayIDused {
			return multiaddr
		}
	}
	return OwnRelayAddrFull
}
