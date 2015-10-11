package ws

import (
	"github.com/jonas747/fnet"
	"golang.org/x/net/websocket"
)

func Dial(addr, protocol, origin string) (fnet.Session, error) {
	nativeConn, err := websocket.Dial(addr, protocol, origin)
	if err != nil {
		return fnet.Session{}, err
	}

	wrappedConn := NewWebsocketConn(nativeConn)

	session := fnet.Session{
		Conn: wrappedConn,
		Data: new(fnet.SessionStore),
	}

	return session, nil
}
