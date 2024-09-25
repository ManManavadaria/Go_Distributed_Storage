package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/ManManavadaria/Go_Distributed_Storage/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddr,
		ShakeHands:    p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		ListenAddr:        listenAddr,
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: CASPathTransform,
		Transport:         tcpTransport,
		BootstrapedNodes:  nodes,
	}

	f := NewFileServer(fileServerOpts)

	tcpTransport.TCPTransportOpts.OnPeer = f.OnPeer

	return f
}

func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")

	go func() {
		log.Fatal(s1.Start())
	}()

	time.Sleep(3 * time.Second)
	go s2.Start()
	time.Sleep(3 * time.Second)

	data := bytes.NewReader([]byte("testing"))

	if err := s2.StoreData("secret", data); err != nil {
		fmt.Println("error : ", err)
	}

	select {}
}
