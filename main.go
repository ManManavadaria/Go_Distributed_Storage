package main

import (
	"fmt"
	"log"

	"github.com/ManManavadaria/Go_Distributed_Storage/p2p"
)

func OnPeer(p2p.Peer) error {
	fmt.Println("Peer is online")
	return nil
}

func main() {

	tcpopts := p2p.TCPTransportOpts{
		ListenAddress: ":3000",
		Decoder:       p2p.DefaultDecoder{},
		ShakeHands:    p2p.NOPHandshakeFunc,
		OnPeer:        OnPeer,
	}

	tr := p2p.NewTCPTransport(tcpopts)

	go func() {
		for rpc := range tr.Consume() {
			// Process RPC here
			fmt.Println("Received RPC:", string(rpc.Payload))
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("hello server")

	select {}
}
