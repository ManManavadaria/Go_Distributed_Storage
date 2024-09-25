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
	Wg       *sync.WaitGroup
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
		Wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
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
	rpc := RPC{}

	for {
		err = t.TCPTransportOpts.Decoder.Decode(conn, &rpc)
		// if err == &net.OpError{}{

		// }
		if err != nil {
			fmt.Println("error after decode : ", err)
			return
		}
		rpc.From = conn.RemoteAddr()

		peer.Wg.Add(1)
		fmt.Println("Waiting till stream is done reading")
		// fmt.Println("sending the data in rpcch : ", rpc)
		t.rpcch <- rpc
		peer.Wg.Wait()
		fmt.Println("Stream done continuinf nomal read loop")
	}
}
