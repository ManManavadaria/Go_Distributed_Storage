package p2p

import (
	"errors"
	"fmt"
	"log"
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

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.conn.Write(b)
	return err
}

func (p *TCPPeer) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}

func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)
	return nil
}

func (t *TCPTransport) ListenAndAccept() {
	var err error

	t.listener, err = net.Listen("tcp", t.TCPTransportOpts.ListenAddress)
	if errors.Is(err, net.ErrClosed) {
		fmt.Println("Network connection closed")
	}
	if err != nil {
		fmt.Println(err)
	}

	go t.startAcceptLoop()

	log.Printf("TCP transport listening on %s", t.TCPTransportOpts.ListenAddress)
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP eccept error: %s \n", err)
		}
		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error

	defer func() {
		fmt.Printf("Dropping the peer connection: %v", err)
		conn.Close()
	}()
	peer := NewTCPPeer(conn, outbound)

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
