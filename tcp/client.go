package tcp

import (
	"github.com/jonas747/fnet"
	"net"
)

func Dial(addr string) (fnet.Session, error) {
	nativeConn, err := net.Dial("tcp", addr)
	if err != nil {
		return fnet.Session{}, err
	}

	wrappedConn := NewTCPConn(nativeConn)
	session := fnet.Session{
		Conn: wrappedConn,
		Data: new(fnet.SessionStore),
	}

	return session, nil
}
