package p2p

import "net"

type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

type Transport interface {
	Addr() string
	ListenAndAccept()
	Consume() <-chan RPC
	Close() error
	Dial(string) error
}
