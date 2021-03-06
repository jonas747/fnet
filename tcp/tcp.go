package tcp

import (
	"github.com/jonas747/fnet"
	"io"
	"net"
	"time"
)

type TCPListner struct {
	Engine    *fnet.Engine
	Addr      string
	Listening bool
	StopChan  chan chan bool
}

// Implements fnet.Listener.Listen
func (t *TCPListner) Listen() error {
	t.Listening = true

	defer func() {
		t.Listening = false
	}()

	listener, err := net.Listen("tcp", t.Addr)
	if err != nil {
		return err
	}

	for {
		// Check if we should stop
		select {
		case rChan := <-t.StopChan:
			rChan <- true
			return nil
		default:
			// Continue normally
		}

		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		wrappedConn := NewTCPConn(conn)
		session := fnet.Session{
			Conn: wrappedConn,
			Data: new(fnet.SessionStore),
		}
		go t.Engine.HandleConn(session)
	}
}

// Implements fnet.Listener.IsListening
func (t *TCPListner) IsListening() bool {
	return t.Listening
}

func (t *TCPListner) Stop() error {
	rChan := make(chan bool)
	t.StopChan <- rChan
	ok := <-rChan
	if !ok {
		return fnet.ErrCantStopListener
	}
	return nil
}

type TCPConn struct {
	sessionStore *fnet.SessionStore
	conn         net.Conn

	writeChan   chan []byte
	stopWriting chan bool
	writing     bool

	isOpen bool
}

func NewTCPConn(c net.Conn) fnet.Connection {
	store := &fnet.SessionStore{make(map[string]interface{})}
	conn := TCPConn{
		sessionStore: store,
		conn:         c,
		writeChan:    make(chan []byte),
		stopWriting:  make(chan bool),
		isOpen:       true,
	}
	return &conn
}

// Implements Connection.Send([]byte)
func (t *TCPConn) Send(b []byte) error {
	if !t.isOpen {
		return fnet.ErrConnClosed
	}
	after := time.After(time.Duration(5) * time.Second) // Time out
	select {
	case t.writeChan <- b:
		return nil
	case <-after:
		t.Close()
		return fnet.ErrTimeout
	}
}

func (t *TCPConn) Read(buf []byte) error {
	_, err := io.ReadFull(t.conn, buf)
	return err
}

// Implements Connection.Kind() string
func (t *TCPConn) Kind() string {
	return "websocket"
}

// Implements Connection.Close()
func (t *TCPConn) Close() {
	if t.isOpen {
		t.isOpen = false
		t.conn.Close()
	}
	if t.writing {
		t.stopWriting <- true
	}
}

func (t *TCPConn) Open() bool {
	return t.isOpen
}

func (t *TCPConn) GetSessionData() *fnet.SessionStore {
	return t.sessionStore
}

// Implements Connection.Run()
func (t *TCPConn) Run() {
	// Launch the write goroutine
	go t.writer()
}

func (t *TCPConn) writer() {
	t.writing = true
	defer func() {
		t.writing = false
	}()
	for {
		select {
		case m := <-t.writeChan:
			_, err := t.conn.Write(m)
			if err != nil {
				break
			}
		case <-t.stopWriting:
			return
		}
	}
}
