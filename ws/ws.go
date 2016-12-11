package ws

import (
	"errors"
	"github.com/jonas747/fnet"
	"golang.org/x/net/websocket"
	"net/http"
	"strings"
	"sync"
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
		session := fnet.Session{
			Conn: conn,
			Data: new(fnet.SessionStore),
		}
		w.Engine.HandleConn(session)
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
	sync.Mutex
	isOpen bool
}

func NewWebsocketConn(c *websocket.Conn) fnet.Connection {
	store := &fnet.SessionStore{
		Data: make(map[string]interface{}),
	}
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
	after := time.After(time.Duration(60) * time.Second) // Time out
	select {
	case w.writeChan <- b:
		return nil
	case <-after:
		w.Close()
		return errors.New("Timed out sending payload to writechan")
	}
}

func (w *WebsocketConn) Read(buf []byte) error {
	if !w.isOpen {
		return errors.New("Can't read from closed connection")
	}

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
	w.Lock()
	if w.writing {
		w.writing = false
		w.Unlock()
		w.stopWriting <- true
	} else {
		w.Unlock()
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
		w.Lock()
		w.writing = false
		w.Unlock()
		//close(w.stopWriting)
	}()
	for {
		select {
		case m := <-w.writeChan:
			err := websocket.Message.Send(w.conn, m)
			if err != nil {
				return
			}
		case <-w.stopWriting:
			return
		}
	}
}

func (w *WebsocketConn) IP() string {
	addr := w.conn.Request().RemoteAddr
	split := strings.Split(addr, ":")
	return split[0]
}
