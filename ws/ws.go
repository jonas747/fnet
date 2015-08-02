package ws

import (
	"errors"
	"fmt"
	"github.com/jonas747/fnet"
	"golang.org/x/net/websocket"
	"net/http"
	"time"
)

type WebsocketListener struct {
	Engine    *fnet.Engine
	Addr      string
	Listening bool
}

// Implements fnet.Listener.Listen
func (w *WebsocketListener) Listen() error {
	handler := func(ws *websocket.Conn) {
		conn := NewWebsocketConn(ws)
		fmt.Println("Received new connection")
		w.Engine.HandleConn(conn)
	}

	server := websocket.Server{Handler: handler}

	http.Handle("/", server)
	w.Listening = true
	err := http.ListenAndServe(w.Addr, nil)
	w.Listening = false
	if err != nil {
		return err
	}
	return nil
}

// Implements fnet.Listener.IsListening
func (w *WebsocketListener) IsListening() bool {
	return w.Listening
}

func (w *WebsocketListener) Stop() error {
	// TODO, need to use a stoppable http server
	return nil
}

type WebsocketConn struct {
	sessionStore *fnet.SessionStore
	conn         *websocket.Conn

	writeChan   chan []byte
	stopWriting chan bool
	writing     bool

	isOpen bool
}

func NewWebsocketConn(c *websocket.Conn) fnet.Connection {
	store := &fnet.SessionStore{make(map[string]interface{})}
	conn := WebsocketConn{
		sessionStore: store,
		conn:         c,
		writeChan:    make(chan []byte),
		stopWriting:  make(chan bool),
		isOpen:       true,
	}
	return &conn
}

// Implements Connection.Send([]byte)
func (w *WebsocketConn) Send(b []byte) error {
	if !w.isOpen {
		return errors.New("Cannot call WebsocketConn.Send() on a closed connection")
	}
	after := time.After(time.Duration(5) * time.Second) // Time out
	select {
	case w.writeChan <- b:
		return nil
	case <-after:
		w.isOpen = false
		w.Close()
		return errors.New("Timed out sending payload to writechan")
	}
}

func (w *WebsocketConn) Read(buf []byte) error {
	_, err := w.conn.Read(buf)
	return err
}

// Implements Connection.Kind() string
func (w *WebsocketConn) Kind() string {
	return "websocket"
}

// Implements Connection.Close()
func (w *WebsocketConn) Close() {
	if w.isOpen {
		w.isOpen = false
		w.conn.Close()
	}
	if w.writing {
		w.stopWriting <- true
	}
}

func (w *WebsocketConn) Open() bool {
	return w.isOpen
}

func (w *WebsocketConn) GetSessionData() *fnet.SessionStore {
	return w.sessionStore
}

// Implements Connection.Run()
func (w *WebsocketConn) Run() {
	// Launch the write goroutine
	go w.writer()
}

// Writes messages from WebsocketConn.writeChan, Which is used by WebsocketConn.Write([]byte)
func (w *WebsocketConn) writer() {
	w.writing = true
	defer func() {
		w.writing = false
	}()
	for {
		select {
		case m := <-w.writeChan:
			err := websocket.Message.Send(w.conn, m)
			if err != nil {
				break
			}
		case <-w.stopWriting:
			return
		}
	}
}
