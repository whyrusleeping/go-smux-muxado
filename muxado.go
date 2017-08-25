package peerstream_muxado

import (
	"net"
	"time"

	muxado "github.com/inconshreveable/muxado"
	smux "github.com/libp2p/go-stream-muxer"
)

// stream implements smux.Stream using a ss.Stream
type stream struct {
	ms muxado.Stream
}

func (s *stream) muxadoStream() muxado.Stream {
	return s.ms
}

func (s *stream) Read(buf []byte) (int, error) {
	return s.ms.Read(buf)
}

func (s *stream) Write(buf []byte) (int, error) {
	return s.ms.Write(buf)
}

func (s *stream) Close() error {
	return s.ms.CloseWrite()
}

func (s *stream) Reset() error {
	return s.ms.Close()
}

func (s *stream) SetDeadline(t time.Time) error {
	return s.ms.SetDeadline(t)
}

func (s *stream) SetReadDeadline(t time.Time) error {
	return s.ms.SetReadDeadline(t)
}

func (s *stream) SetWriteDeadline(t time.Time) error {
	return s.ms.SetWriteDeadline(t)
}

var _ smux.Stream = (*stream)(nil)

// Conn is a connection to a remote peer.
type conn struct {
	ms muxado.Session

	closed chan struct{}
}

func (c *conn) muxadoSession() muxado.Session {
	return c.ms
}

func (c *conn) Close() error {
	return c.ms.Close()
}

func (c *conn) IsClosed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

// OpenStream creates a new stream.
func (c *conn) OpenStream() (smux.Stream, error) {
	s, err := c.ms.OpenStream()
	if err != nil {
		return nil, err
	}

	return &stream{ms: s}, nil
}

// AcceptStream accepts a stream opened by the other side.
func (c *conn) AcceptStream() (smux.Stream, error) {
	s, err := c.ms.AcceptStream()
	if err != nil {
		return nil, err
	}
	return &stream{ms: s}, nil
}

type transport muxado.Config

// Transport is a go-peerstream transport that constructs
// spdystream-backed connections.
var Transport = &transport{
	AcceptBacklog: 2048,
	MaxWindowSize: 256 * 1 << 10,
}

func (t *transport) NewConn(nc net.Conn, isServer bool) (smux.Conn, error) {
	var s muxado.Session
	if isServer {
		s = muxado.Server(nc, (*muxado.Config)(t))
	} else {
		s = muxado.Client(nc, (*muxado.Config)(t))
	}
	cl := make(chan struct{})
	go func() {
		s.Wait()
		close(cl)
	}()
	return &conn{ms: s, closed: cl}, nil
}
