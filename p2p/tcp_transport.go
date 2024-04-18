package p2p

import (
	"fmt"
	"net"
	"sync"
)

type TCPTransportOpts struct {
	ListenAddress string
	Decoder       Decoder
	ShakeHands    HandshakeFunc
}

type TCPTransport struct {
	TCPTransportOpts TCPTransportOpts
	listener         net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

type TCPPeer struct {
	conn     net.Conn
	outbound bool
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
	}
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.TCPTransportOpts.ListenAddress)
	if err != nil {
		return err
	}
	go t.startAcceptLoop()
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP eccept error: %s \n", err)
		}
		go t.handleConn(conn)
	}
}

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	if err := t.TCPTransportOpts.ShakeHands(peer); err != nil {
		conn.Close()
		return
	}
	// msg := &Message{}
	// for {
	// if err := t.TCPTransportOpts.Decoder.Decode(conn, msg); err != nil {
	fmt.Println("new conn", peer)
	// }

	// }
}
