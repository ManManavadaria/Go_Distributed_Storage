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
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts TCPTransportOpts
	listener         net.Listener
	rpcch            chan RPC

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
		rpcch:            make(chan RPC),
	}
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
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

func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error

	defer func() {
		fmt.Printf("Dropping the peer connection: %v", err)
		conn.Close()
	}()
	peer := NewTCPPeer(conn, true)

	if err := t.TCPTransportOpts.ShakeHands(peer); err != nil {
		conn.Close()
		return
	}

	if err = t.TCPTransportOpts.OnPeer(peer); err != nil {
		return
	}

	//read loop
	rpc := RPC{}

	for {
		err = t.TCPTransportOpts.Decoder.Decode(conn, &rpc)
		// if err == &net.OpError{}{

		// }
		if err != nil {
			return
		}
		rpc.From = conn.RemoteAddr()

		t.rpcch <- rpc
	}
}
