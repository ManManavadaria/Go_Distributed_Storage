package p2p

import "net"

type RPC struct {
	From    net.Addr
	Payload []byte
	Stream  bool
}

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)
