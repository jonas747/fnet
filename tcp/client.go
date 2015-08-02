package tcp

import (
	"github.com/jonas747/fnet"
	"net"
)

func Dial(addr string) (fnet.Connection, error) {
	nativeConn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewTCPConn(nativeConn), nil
}
