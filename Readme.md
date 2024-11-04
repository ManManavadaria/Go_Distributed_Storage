# Go Distributed File Storage System

A robust distributed file storage system implemented in Go, featuring peer-to-peer networking, encryption, and distributed storage capabilities.

## Overview

This project implements a sophisticated distributed file storage system where multiple nodes work together to store, retrieve, and manage files across a network. Each node functions as both client and server, creating a true peer-to-peer network with built-in encryption for security.

### Key Features

- 🔗 Peer-to-peer architecture with dynamic node discovery
- 🔒 AES encryption for secure file storage
- 📁 Content-addressable storage using SHA-1 and MD5 hashing
- 🔄 Automatic file distribution across network nodes
- 🚀 Custom TCP-based transport layer
- 🔍 Distributed file retrieval with streaming support
- ⚡ Non-blocking concurrent operations
- 🔐 Secure file removal across the network

## System Architecture

The project is organized into several core components:

### Main Application (`main.go`)
- Initializes server instances with customizable configurations
- Processes user commands through an interactive CLI
- Manages node bootstrapping and peer connections
- Implements command validation and processing

### File Server (`server.go`)
- Handles core distributed storage operations
- Manages peer-to-peer message routing
- Implements file streaming and chunked transfer
- Coordinates network-wide file operations

### Storage Layer (`store.go`)
- Provides content-addressable storage
- Manages local file system operations
- Implements path transformation and file handling
- Supports atomic file operations

### Cryptography (`crypto.go`)
- Implements AES-CTR encryption
- Provides secure key generation
- Handles stream-based encryption/decryption
- Includes MD5 hashing for file identification

### P2P Networking (`p2p/`)
- **TCP Transport**: Custom implementation for peer communication
- **Message Handling**: Defines RPC structures and protocols
- **Stream Processing**: Supports large file transfers
- **Connection Management**: Handles peer lifecycle

## Installation

### Prerequisites
- Go 1.16 or higher
- Git

### Setup
```bash
# Clone the repository
git clone https://github.com/ManManavadaria/Go_Distributed_Storage.git
cd Go_Distributed_Storage

# Install dependencies
go mod download
```

## Usage

### Build
1. For Windows:
```bash
go build -o dfss-build.exe
```

2. For other OS:
```bash
go build -o dfss-build
```

### Starting the Network

1. Start the first node:
```bash
./dfss-build.exe -port :3000 -nodes :4000,:5000
```

2. Start additional nodes:
```bash
./dfss-build.exe -port :4000 -nodes :3000,:5000
./dfss-build.exe -port :5000 -nodes :4000,:3000
```

### Command Interface

The system provides an interactive command interface with the following format:
```
action,key,content
```

Available commands:

1. **Write a file**:
```
write,filename,content
```

2. **Read a file**:
```
read,filename
```

3. **Remove a file**:
```
remove,filename
```

### Implementation Details

#### File Storage Mechanism

Files are stored using a content-addressable system with a sophisticated path transformation:

```go
func CASPathTransform(key string) PathKey {
    hash := sha1.Sum([]byte(key))
    hashStr := hex.EncodeToString(hash[:])
    blockSize := 5
    sliceLen := len(hashStr) / blockSize
    paths := make([]string, sliceLen)
    
    for i := 0; i < sliceLen; i++ {
        from, to := i*blockSize, (i*blockSize)+blockSize
        paths[i] = hashStr[from:to]
    }
    
    return PathKey{
        PathName: strings.Join(paths, "/"),
        FileName: hashStr,
    }
}
```

#### Encryption System

The system uses AES-CTR mode encryption with streaming support:

```go
func copyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
    block, err := aes.NewCipher(key)
    // ... encryption setup
    stream := cipher.NewCTR(block, iv)
    return copyStream(stream, block.BlockSize(), src, dst)
}
```

#### Network Communication

Messages between nodes are handled through a custom RPC system:

```go
type Message struct {
    From    string
    Payload any
}

type MessageStoreFile struct {
    Key  string
    Size int
}
```

## Features in Detail

### Automatic Node Discovery
- Nodes automatically connect to existing peers
- Dynamic peer management with concurrent connection handling
- Graceful connection lifecycle management

### Secure File Operations
- End-to-end encryption for all stored files
- Secure key generation and management
- Stream-based encryption for efficient memory usage

### Distributed Storage
- Content-addressable storage with SHA-1 hashing
- Automatic file replication across nodes
- Concurrent file operations handling

### Error Handling
- Robust error management for network operations
- Graceful degradation on node failures
- Comprehensive error reporting

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.