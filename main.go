package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"

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

type Command struct {
	Action  string
	Key     string
	Content string
}

func main() {

	port := flag.String("port", "", "Server port address")
	nodes := flag.String("nodes", "", "Remote nodes to connect the current node")

	flag.Parse()

	validatePortAddr(*port)
	nodeList := extractAndValidateNodes(*nodes)

	commandChan := make(chan Command)
	doneProcess := make(chan bool)

	fmt.Printf("\n\033[34mNode Initialization and Bootstrap Process =======>\033[0m\n")
	s := makeServer(*port, nodeList...)

	go func() {
		log.Fatal(s.Start())
	}()

	go processCommands(s, commandChan, doneProcess)

	for {
		var input string
		fmt.Print("Enter command (format: action,key,content): ")
		fmt.Scanln(&input)

		parts := strings.SplitN(input, ",", 3)
		if len(parts) < 2 {
			fmt.Println("Invali commadnd format. Please use: action,key,content (content is optional for write)")
			continue
		}

		action := strings.TrimSpace(parts[0])
		key := strings.TrimSpace(parts[1])
		content := ""
		if action == "write" && len(parts) == 3 {
			content = strings.TrimSpace(parts[2])
		}
		commandChan <- Command{Action: action, Key: key, Content: content}
		<-doneProcess
	}
}

func processCommands(s *FileServer, commandChan chan Command, done chan bool) {
	for command := range commandChan {
		switch command.Action {
		case "write":
			fmt.Printf("\n\033[34mWriting File =======>\033[0m\n")
			data := bytes.NewReader([]byte(command.Content))
			if err := s.store(command.Key, data); err != nil {
				fmt.Println("error : ", err)
			}

		case "read":
			fmt.Printf("\n\033[34mReading File =======>\033[0m\n")
			_, r, err := s.Get(command.Key)
			if err != nil {
				log.Fatal(err)
			}
			b, err := io.ReadAll(r)
			if err != nil {
				log.Fatal(err)
			}
			if rc, ok := r.(io.ReadCloser); ok {
				rc.Close()
			}
			fmt.Printf("\nContent of file %s: %s\n", command.Key, string(b))

		case "remove":
			fmt.Printf("\n\033[34mRemoving File =======>\033[0m\n")
			if err := s.Remove(command.Key); err != nil {
				log.Fatal(err)
			}
		default:
			fmt.Println("Invalid command")
		}
		done <- true
	}
}

func validatePortAddr(port string) {
	if len(port) == 0 && !strings.HasPrefix(port, ":") {
		log.Fatal("Invalid port argument.", port)
	}
}

func extractAndValidateNodes(nodes string) []string {
	if len(nodes) == 0 {
		log.Fatal("Invalid nodes")
	}
	nodeList := strings.Split(nodes, ",")
	for i, node := range nodeList {
		nodeList[i] = strings.TrimSpace(node)
		validatePortAddr(nodeList[i])
	}
	return nodeList
}
