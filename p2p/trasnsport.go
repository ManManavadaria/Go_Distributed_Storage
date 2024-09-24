package p2p

import "net"

type Peer interface {
	Close() error
	RemoteAddr() net.Addr
	Send([]byte) error
}

type Transport interface {
	ListenAndAccept()
	Consume() <-chan RPC
	Close() error
	Dial(string) error
}
