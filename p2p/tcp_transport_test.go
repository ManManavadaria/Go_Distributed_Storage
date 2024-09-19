package p2p

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestTCPTransport(t *testing.T) {
	tcpopts := TCPTransportOpts{
		ListenAddress: ":7000",
		Decoder:       GOBDecoder{},
		ShakeHands:    NOPHandshakeFunc,
	}

	tr := NewTCPTransport(tcpopts)

	assert.Equal(t, tr.TCPTransportOpts.ListenAddress, ":7000")
}
