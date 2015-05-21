package ws

import (
	"code.google.com/p/go.net/websocket"
	"github.com/jonas747/fnet"
)

func Dial(addr, protocol, origin string) (fnet.Connection, error) {
	nativeConn, err := websocket.Dial(addr, protocol, origin)
	if err != nil {
		return nil, err
	}

	wrappedConn := NewWebsocketConn(nativeConn)
	return wrappedConn, nil
}
