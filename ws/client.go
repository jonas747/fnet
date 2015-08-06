package ws

import (
	"github.com/jonas747/fnet"
	"golang.org/x/net/websocket"
)

func Dial(addr, protocol, origin string) (fnet.Connection, error) {
	nativeConn, err := websocket.Dial(addr, protocol, origin)
	if err != nil {
		return nil, err
	}

	wrappedConn := NewWebsocketConn(nativeConn)
	return wrappedConn, nil
}
