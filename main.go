package main

import (
	"fmt"
	"log"

	"github.com/ManManavadaria/Go_Distributed_Storage/p2p"
)

func main() {
	tcpopts := p2p.TCPTransportOpts{
		ListenAddress: ":3000",
		Decoder:       p2p.GOBDecoder{},
		ShakeHands:    p2p.NOPHandshakeFunc,
	}

	tr := p2p.NewTCPTransport(tcpopts)

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("hello server")

	select {}
}
