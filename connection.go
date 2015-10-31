package fnet

import (
	"sync"
)

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
	Data  map[string]interface{}
	Mutex sync.RWMutex
}

func (s *SessionStore) Set(key string, val interface{}) {
	s.Mutex.Lock()
	if s.Data == nil {
		s.Data = make(map[string]interface{})
	}
	s.Data[key] = val
	s.Mutex.Unlock()
}

func (s *SessionStore) Get(key string) (val interface{}, exists bool) {
	s.Mutex.RLock()
	if s.Data == nil {
		s.Data = make(map[string]interface{})
	}

	value, exists := s.Data[key]
	s.Mutex.RUnlock()
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

func (s *SessionStore) GetBool(key string) (val bool, exists bool) {
	raw, exists := s.Get(key)
	if exists {
		val, _ = raw.(bool)
	}
	return
}

func (s *SessionStore) GetInt(key string) (val int, exists bool) {
	raw, exists := s.Get(key)
	if exists {
		val, _ = raw.(int)
	}
	return
}
