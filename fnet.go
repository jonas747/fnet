package fnet

import (
	"errors"
	"reflect"
)

var (
	ErrCantStopListener   = errors.New("Unable to stop listener")
	ErrSliceLengthsDiffer = errors.New("Slice lengths differ")
	ErrConnClosed         = errors.New("Connection closed")
	ErrTimeout            = errors.New("Timed out")
	ErrNoHandlerFound     = errors.New("No Handler found")
)

// Listener is a interface for listening for incoming connections
type Listener interface {
	Listen() error     // Listens for incoming connections
	IsListening() bool // Returns wether this listener is listening or not
	Stop() error       // Stops listening for incoming connections and
}

// Struct which represents an event
type Event struct {
	Id   int32
	Data reflect.Value
}

// Struct which represents a event handler
type Handler struct {
	CallBack interface{}
	Event    int32
	DataType reflect.Type
}

func NewHandler(callback interface{}, evt int32) (Handler, error) {
	err := validateCallback(callback)
	if err != nil {
		return Handler{}, err
	}

	t := reflect.TypeOf(callback)
	var dType reflect.Type
	if t.NumIn() == 2 {
		dType = reflect.TypeOf(callback).In(1) // Get the type of the first parameter for later use
	}

	return Handler{
		CallBack: callback,
		Event:    evt,
		DataType: dType,
	}, nil
}

// Panics if an error occurs instead of returning it
func NewHandlerSafe(callback interface{}, evt int32) Handler {
	err := validateCallback(callback)
	if err != nil {
		panic(err)
	}

	t := reflect.TypeOf(callback)

	var dType reflect.Type
	if t.NumIn() == 2 {
		dType = reflect.TypeOf(callback).In(1) // Get the type of the first parameter for later use
	}

	return Handler{
		CallBack: callback,
		Event:    evt,
		DataType: dType,
	}
}

func validateCallback(callback interface{}) error {
	t := reflect.TypeOf(callback)
	if t.Kind() != reflect.Func {
		return errors.New("Callback not a function")
	}
	return nil
}
