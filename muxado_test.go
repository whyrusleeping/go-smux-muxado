package peerstream_muxado

import (
	"testing"

	test "github.com/libp2p/go-stream-muxer/test"
)

func TestMuxadoTransport(t *testing.T) {
	test.SubtestAll(t, Transport)
}
