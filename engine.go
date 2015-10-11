package fnet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

// The networking engine. Holds togheter all the connections and handlers
type Engine struct {
	Encoder     Encoder // The encoder/decoder to use
	OnConnOpen  func(Session)
	OnConnClose func(Session)

	registerSession   chan Session // Channel for registering new connections
	unregisterSession chan Session // Channel for unregistering connections
	broadcastChan     chan []byte  // Channel for broadcasting messages to all connections

	listeners []Listener        // Slice Containing all listeners
	handlers  map[int32]Handler // Map with all the event handlers, their id's as keys
	sessions  map[Session]bool  // Map containing all conncetions
	ErrChan   chan error
}

func DefaultEngine() *Engine {
	return &Engine{
		registerSession:   make(chan Session),
		unregisterSession: make(chan Session),
		broadcastChan:     make(chan []byte),
		listeners:         make([]Listener, 0),
		handlers:          make(map[int32]Handler),
		sessions:          make(map[Session]bool),
		ErrChan:           make(chan error),
		Encoder:           ProtoEncoder{},
	}
}

func (e *Engine) Broadcast(msg []byte) {
	e.broadcastChan <- msg
}

// Adds a listener and make it start listening for incoming connections
func (e *Engine) AddListener(listener Listener) {
	e.listeners = append(e.listeners, listener)
	if !listener.IsListening() {
		go func() {
			err := listener.Listen()
			e.ErrChan <- err
		}()
	}
}

// Handles connections
func (e *Engine) HandleConn(session Session) {
	session.Conn.Run()

	e.registerSession <- session
	if e.OnConnOpen != nil {
		e.OnConnOpen(session)
	}

	for {
		err := e.readMessage(session)
		if err != nil {
			fmt.Println("Error: ", err)
			if err != ErrNoHandlerFound {
				break
			}
		}
	}

	session.Conn.Close()
	e.unregisterSession <- session
	if e.OnConnClose != nil {
		e.OnConnClose(session)
	}
}

func (e *Engine) readMessage(session Session) error {
	conn := session.Conn

	// start with receving the evt id and payload length
	header := make([]byte, 8)
	err := conn.Read(header)
	if err != nil {
		return err
	}
	evtId, pl, err := readHeader(header)
	if err != nil {
		return err
	}
	payload := make([]byte, pl)
	if pl > 0 {
		err = conn.Read(payload)
		if err != nil {
			return err
		}
	} else {
		//fmt.Println("No payload!")
	}
	return e.handleMessage(evtId, payload, session)
}

func readHeader(header []byte) (evtId int32, payloadLength int32, err error) {
	buf := bytes.NewReader(header)

	err = binary.Read(buf, binary.LittleEndian, &evtId)
	if err != nil {
		return
	}

	err = binary.Read(buf, binary.LittleEndian, &payloadLength)
	if err != nil {
		return
	}
	return
}

// Retrieves the event id, decodes the data and calls the callback
func (e *Engine) handleMessage(evtId int32, payload []byte, seesion Session) error {
	handler, found := e.handlers[evtId]
	if !found {
		return ErrNoHandlerFound
	}

	var args = make([]reflect.Value, 0)
	sesisonVal := reflect.ValueOf(seesion)
	args = append(args, sesisonVal)
	if len(payload) > 0 {
		decoded := reflect.New(handler.DataType).Interface() // We use reflect to unmarshal the data into the appropiate typewww
		err := e.Encoder.Unmarshal(payload, decoded)
		if err != nil {
			return err
		}
		decVal := reflect.Indirect(reflect.ValueOf(decoded)) // decoded is a pointer, so we get the value it points to
		args = append(args, decVal)
	}
	// ready the function
	funcVal := reflect.ValueOf(handler.CallBack)
	resp := funcVal.Call(args) // Call it
	if len(resp) == 0 {
		return nil
	}

	// Todo, allow the handlers to return stuff to send
	// for a simple request response type structure

	return nil
}

// Adds a handler
func (e *Engine) AddHandler(handler Handler) {
	e.handlers[handler.Event] = handler
}

// Adds multiple handlers
func (e *Engine) AddHandlers(handlers ...Handler) {
	for _, v := range handlers {
		e.AddHandler(v)
	}
}

// Listen for messages on all the channels
func (e *Engine) ListenChannels() {
	for {
		select {
		case d := <-e.registerSession: //Register a connection
			e.sessions[d] = true
		case d := <-e.unregisterSession: //Unregister a connection
			delete(e.sessions, d)
		case msg := <-e.broadcastChan: //Broadcast a message to all connections
			for sess := range e.sessions {
				err := sess.Conn.Send(msg)
				if err != nil {
					e.ErrChan <- err
				}
			}
		}
	}
}

func (e *Engine) NumClients() int {
	return len(e.sessions)
}

func (e *Engine) CreateWireMessage(evtId int32, data interface{}) ([]byte, error) {
	// Encode the message itself
	encoded := make([]byte, 0)
	if data != nil {
		e, err := e.Encoder.Marshal(data)
		if err != nil {
			return make([]byte, 0), err
		}
		encoded = e
	}

	// Create a new buffer, stuff the event id and the encoded message in it
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, evtId)
	if err != nil {
		return make([]byte, 0), err
	}

	// Add the length to the buffer
	length := len(encoded)
	err = binary.Write(buffer, binary.LittleEndian, int32(length))
	if err != nil {
		return make([]byte, 0), err
	}

	// Then the actual payload, if any
	if len(encoded) > 0 {
		_, err = buffer.Write(encoded)
		if err != nil {
			return make([]byte, 0), err
		}

	}

	unread := buffer.Bytes()
	return unread, nil
}

func (e *Engine) CreateAndSend(session Session, evtId int32, data interface{}) error {
	wireMessage, err := e.CreateWireMessage(evtId, data)
	if err != nil {
		return err
	}

	return session.Conn.Send(wireMessage)
}

func (e *Engine) CreateAndBroadcast(evtId int32, data interface{}) error {
	wireMessage, err := e.CreateWireMessage(evtId, data)
	if err != nil {
		return err
	}
	e.Broadcast(wireMessage)
	return nil
}
