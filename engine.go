package fnet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

// The networking engine. Holds togheter all the connections and handlers
type Engine struct {
	ConnCloseChan   chan Connection
	EmitConnOnClose bool
	Encoder         Encoder // The encoder/decoder to use

	registerConn   chan Connection // Channel for registering new connections
	unregisterConn chan Connection // Channel for unregistering connections
	broadcastChan  chan []byte     // Channel for broadcasting messages to all connections

	listeners   []Listener          // Slice Containing all listeners
	handlers    map[int32]Handler   // Map with all the event handlers, their id's as keys
	connections map[Connection]bool // Map containing all conncetions
	ErrChan     chan error
}

func DefaultEngine() *Engine {
	return &Engine{
		ConnCloseChan:  make(chan Connection),
		registerConn:   make(chan Connection),
		unregisterConn: make(chan Connection),
		broadcastChan:  make(chan []byte),
		listeners:      make([]Listener, 0),
		handlers:       make(map[int32]Handler),
		connections:    make(map[Connection]bool),
		ErrChan:        make(chan error),
		Encoder:        ProtoEncoder{},
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
func (e *Engine) HandleConn(conn Connection) {
	conn.Run()
	e.registerConn <- conn
	for {
		err := e.readMessage(conn)
		if err != nil {
			fmt.Println("Error: ", err)
			if err != ErrNoHandlerFound {
				break
			}
		}
	}

	conn.Close()
	e.unregisterConn <- conn
}

func (e *Engine) readMessage(conn Connection) error {
	// start with receving the evt id and payload length
	header := make([]byte, 8)
	fmt.Println("Waiting on header...")
	err := conn.Read(header)
	fmt.Println("Received header! prcoessing it...")
	if err != nil {
		return err
	}
	evtId, pl, err := readHeader(header)
	if err != nil {
		return err
	}
	fmt.Println(evtId, pl)

	payload := make([]byte, pl)
	if pl > 0 {
		fmt.Println("Waiting on payload")
		if pl != 0 {
			err = conn.Read(payload)
			if err != nil {
				return err
			}
		}
	} else {
		fmt.Println("No payload...")
	}
	fmt.Println("Received payload! Processing it...")
	return e.handleMessage(evtId, payload, conn)
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
func (e *Engine) handleMessage(evtId int32, payload []byte, conn Connection) error {
	handler, found := e.handlers[evtId]
	if !found {
		return ErrNoHandlerFound
	}

	var args = make([]reflect.Value, 0)
	connVal := reflect.ValueOf(conn)
	args = append(args, connVal)
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

	return nil
}

// Adds a handler
func (e *Engine) AddHandler(handler Handler) {
	e.handlers[handler.Event] = handler
}

func (e *Engine) AddHandlers(handlers ...Handler) {
	for _, v := range handlers {
		e.AddHandler(v)
	}
}

func (e *Engine) ListenChannels() {
	for {
		select {
		case d := <-e.registerConn: //Register a connection
			e.connections[d] = true
		case d := <-e.unregisterConn: //Unregister a connection
			delete(e.connections, d)
			if e.EmitConnOnClose {
				e.ConnCloseChan <- d
			}
		case msg := <-e.broadcastChan: //Broadcast a message to all connections
			for conn := range e.connections {
				err := conn.Send(msg)
				if err != nil {
					e.ErrChan <- err
				}
			}
		}
	}
}

func (e *Engine) NumClients() int {
	return len(e.connections)
}
