package fnet

import ()

type Connection interface {
	Send([]byte) error             // Sends some data
	Read([]byte) error             // reads data into supplied byte slice
	Kind() string                  // What kind of connection is it (websocket, tcp etc..)
	Close()                        // Closes the connections ands stops all goroutines associated with it
	GetSessionData() *SessionStore // Gets the session datastore
	Run()                          // Starts the writer and reader goroutines
	Open() bool                    // Wether this connection is open ot not
}

type SessionStore struct {
	Data map[string]interface{}
}

func (s *SessionStore) Set(key string, val interface{}) {
	s.Data[key] = val
}

func (s *SessionStore) Get(key string) (val interface{}, exists bool) {
	value, exists := s.Data[key]
	return value, exists
}
