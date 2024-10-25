package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/ManManavadaria/Go_Distributed_Storage/p2p"
)

// makeServer initializes and returns a new FileServer instance.
// It sets up the TCP transport options, encryption key, storage root, and bootstrap nodes.
func makeServer(listenAddr string, nodes ...string) *FileServer {

	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddr,           // Address to listen on
		ShakeHands:    p2p.NOPHandshakeFunc, // No-operation handshake function
		Decoder:       p2p.DefaultDecoder{}, // Default decoder for incoming messages
	}

	// Initialize a new TCP transport with the specified options
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	// Configure FileServer options
	fileServerOpts := FileServerOpts{
		ListenAddr:        listenAddr,              // Address to listen on
		StorageRoot:       listenAddr + "_network", // Directory for storing files
		PathTransformFunc: CASPathTransform,        // Function to transform file paths
		Transport:         tcpTransport,            // Transport layer for communication
		BootstrapedNodes:  nodes,                   // Initial peers to connect to
		EncKey:            NewEncryptionKey(),      // Encryption key for securing data
	}

	// Create a new FileServer with the specified options
	f := NewFileServer(fileServerOpts)

	// Assign the OnPeer callback to handle new peer connections
	tcpTransport.TCPTransportOpts.OnPeer = f.OnPeer

	return f
}

func main() {
	actions := flag.String("actions", "write", "Comma-separated actions to perform: read, write, remove")
	fileName := flag.String("file", "secret", "Name of the file to read/write/remove")
	content := flag.String("content", "testing", "Content to write (used for write action)")

	flag.Parse()

	fmt.Printf("\n\033[34mNode Initialization and Bootstrap Process =======>\033[0m\n")

	// Initialize three FileServer instances on different ports
	s1 := makeServer(":3000", "")               // First server with no bootstrap nodes
	s2 := makeServer(":4000", ":3000")          // Second server bootstraps to the first server
	s3 := makeServer(":5000", ":4000", ":3000") // Third server bootstraps to the first and second servers

	// Start the first server in a separate goroutine
	go func() {
		log.Fatal(s1.Start())
	}()

	// Wait for the first server to initialize
	time.Sleep(3 * time.Second)

	// Start the second server in a separate goroutine
	go s2.Start()

	// Wait for the second server to initialize
	time.Sleep(3 * time.Second)

	go func() {
		log.Fatal(s3.Start())
	}()

	time.Sleep(3 * time.Second)

	actionList := strings.Split(*actions, ",")
	for _, action := range actionList {
		action = strings.TrimSpace(action)
		switch action {

		case "write":
			fmt.Printf("\n\033[34mFile Storage Process =======>\033[0m\n")
			data := bytes.NewReader([]byte(*content))

			if err := s2.store(*fileName, data); err != nil {
				fmt.Println("error : ", err)
			}

		case "read":
			fmt.Printf("\n\033[34mFile Reading Process =======>\033[0m\n")
			time.Sleep(time.Second * 3)
			_, r, err := s2.Get(*fileName)
			if err != nil {
				log.Fatal(err)
			}

			// Read all data from the retrieved reader
			b, err := io.ReadAll(r)
			if err != nil {
				log.Fatal(err)
			}

			if rc, ok := r.(io.ReadCloser); ok {
				rc.Close()
			}
			fmt.Printf("\n\033[33mContent of file \033[1m%s\033[0m: \033[32m%s\033[0m\n", *fileName, string(b))

			// To remove the stored files from each nodes
		case "remove":
			fmt.Printf("\n\033[34mFile Deletion Process =======>\033[0m\n")
			time.Sleep(time.Second * 3)
			if err := s2.Remove(*fileName); err != nil {
				log.Fatal(err)
			}
		}
	}
}
