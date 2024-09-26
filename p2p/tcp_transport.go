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
	net.Conn
	outbound bool
	wg       *sync.WaitGroup
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC, 1024),
	}
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}
func (p *TCPPeer) CloseStream() {
	p.wg.Done()
}

func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (t *TCPTransport) Addr() string {
	return t.TCPTransportOpts.ListenAddress
}
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) Dial(addr string) error {

	if len(addr) == 0 {
		return nil
	}

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
		fmt.Println("ListenAndAccept", err)
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

	for {
		rpc := RPC{}
		err = t.TCPTransportOpts.Decoder.Decode(conn, &rpc)
		if err != nil {
			fmt.Println("error after decode : ", err)
			return
		}
		rpc.From = conn.RemoteAddr()

		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("[%s] incoming stream, Waiting...\n", conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Printf("[%s] stream closed, resuming read loop\n", conn.RemoteAddr())
			continue
		}

		t.rpcch <- rpc
	}
}
