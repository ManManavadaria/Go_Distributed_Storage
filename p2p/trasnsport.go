package p2p

import "net"

type Peer interface {
	net.Conn
	Send([]byte) error
}

type Transport interface {
	ListenAndAccept()
	Consume() <-chan RPC
	Close() error
	Dial(string) error
}
