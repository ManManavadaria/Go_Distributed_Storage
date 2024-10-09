# Distributed File Storage System

A decentralized file storage system implemented in Go, demonstrating peer-to-peer networking, encryption, and distributed systems concepts.

## Overview

This project implements a distributed file storage system where multiple nodes can store, retrieve, and remove files across a network. Each node can act both as a client and a server, enabling true peer-to-peer file sharing with built-in encryption for security.

### Key Features

- 🔗 Peer-to-peer architecture
- 🔒 End-to-end encryption of stored files
- 📁 Content-addressable storage
- 🔄 Automatic file replication across nodes
- 🚀 TCP-based transport layer
- 🔍 Distributed file retrieval

## Code Structure

The project consists of several key components:

### Main Application (`main.go`)
- Sets up and initializes multiple file server instances
- Demonstrates file storage, retrieval, and removal operations
- Creates a network of interconnected nodes

### File Server (`server.go`)
- Implements the core file server functionality
- Handles peer connections and message routing
- Manages file operations across the network

### Storage Layer (`store.go`)
- Provides abstractions for file storage and retrieval
- Implements content-addressable storage using SHA-1 hashing
- Manages the local file system for each node

### Cryptography (`crypto.go`)
- Handles encryption and decryption of files
- Implements secure key generation
- Provides utilities for hashing and stream encryption

### P2P Networking Package (`p2p/`)
- **TCP Transport** (`tcp_transport.go`): Implements TCP-based peer-to-peer communication
- **Message Handling** (`message.go`): Defines RPC structures for inter-node communication
- **Encoding** (`encoding.go`): Implements message encoding/decoding
- **Handshake** (`handshake.go`): Defines connection handshake protocols

## Installation

### Prerequisites
- Go 1.16 or higher
- Git

### Steps

1. Clone the repository:
```bash
git clone https://github.com/yourusername/go-distributed-storage.git
cd go-distributed-storage
```

2. Install dependencies:
```bash
go mod download
```

## Usage

### Running the System

1. Start the first node:
```bash
go run . -port 3000
```

2. In separate terminal windows, start additional nodes:
```bash
go run . -port 4000
go run . -port 5000
```

### Basic Operations

#### Storing a File
```go
// Example from main.go
data := bytes.NewReader([]byte("testing"))
err := server.Store("myfile", data)
```

#### Retrieving a File
```go
// Example from main.go
_, reader, err := server.Get("myfile")
if err != nil {
    log.Fatal(err)
}
```

#### Removing a File
```go
// Example from main.go
err := server.Remove("myfile")
```

## Implementation Details

### File Storage
Files are stored using a content-addressable system. The `CASPathTransform` function in `store.go` generates a unique path for each file based on its content:

```go
var CASPathTransform PathTransformFunc = func(key string) PathKey {
    hash := sha1.Sum([]byte(key))
    hashStr := hex.EncodeToString(hash[:])
    // ... path generation logic
}
```

### Encryption
Files are automatically encrypted before storage and decrypted upon retrieval:

```go
func copyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
    block, err := aes.NewCipher(key)
    // ... encryption logic
}
```

### Network Communication
The system uses a custom TCP transport layer for peer-to-peer communication:

```go
type TCPTransport struct {
    TCPTransportOpts
    listener net.Listener
    rpcch    chan RPC
    // ... other fields
}
```
