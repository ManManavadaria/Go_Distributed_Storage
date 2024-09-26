package main

import (
	"fmt"
	"io"
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

	// for i := 0; i < 10; i++ {
	// data := bytes.NewReader([]byte("testing"))

	// if err := s2.store(fmt.Sprintf("secret"), data); err != nil {
	// 	fmt.Println("error : ", err)
	// }

	// 	time.Sleep(time.Millisecond * 500)
	// }

	_, r, err := s2.Get("secret")
	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	if rc, ok := r.(io.ReadCloser); ok {
		defer rc.Close()
	}

	fmt.Println(string(b))
}
