package main

import (
	"log"

	"github.com/ManManavadaria/Go_Distributed_Storage/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddr,
		ShakeHands:    p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	tcpTransport := &p2p.TCPTransport{
		TCPTransportOpts: tcpTransportOpts,
	}

	fileServerOpts := FileServerOpts{
		ListenAddr:        ":3000",
		StorageRoot:       "3000_network",
		PathTransformFunc: CASPathTransform,
		Transport:         tcpTransport,
		BootstrapedNodes:  []string{":4000"},
	}

	f := NewFileServer(fileServerOpts)

	tcpTransport.TCPTransportOpts.OnPeer = f.OnPeer

	return f
}

func main() {
	s1 := makeServer(":4000", "")
	s2 := makeServer(":3000", ":4000")

	go func() {
		log.Fatal(s1.Start())
	}()

	s2.Start()
}
