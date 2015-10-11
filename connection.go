package fnet

import ()

type Connection interface {
	Send([]byte) error // Sends some data
	Read([]byte) error // reads data into supplied byte slice
	Kind() string      // What kind of connection is it (websocket, tcp etc..)
	Close()            // Closes the connections ands stops all goroutines associated with it
	Run()              // Starts the writer and reader goroutines
	Open() bool        // Wether this connection is open ot not
}

type Session struct {
	Data *SessionStore
	Conn Connection
}

type SessionStore struct {
	Data map[string]interface{}
}

func (s *SessionStore) Set(key string, val interface{}) {
	if s.Data == nil {
		s.Data = make(map[string]interface{})
	}
	s.Data[key] = val
}

func (s *SessionStore) Get(key string) (val interface{}, exists bool) {
	if s.Data == nil {
		s.Data = make(map[string]interface{})
	}

	value, exists := s.Data[key]
	return value, exists
}

func (s *SessionStore) GetString(key string) (val string, exists bool) {
	value, exists := s.Get(key)
	if !exists {
		return "", false
	}

	str := value.(string)
	return str, true
}
