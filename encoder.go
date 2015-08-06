package fnet

import (
	"encoding/json"
	"errors"
	"github.com/golang/protobuf/proto"
)

type Encoder interface {
	Marshal(in interface{}) ([]byte, error)
	Unmarshal(data []byte, obj interface{}) error
}

// The standard protocol buffer encoder
type ProtoEncoder struct{}

var ErrNotProtoMessage = errors.New("Proto encoder/decoder reuquires proto.Message to be implemented")

func (p ProtoEncoder) Marshal(in interface{}) ([]byte, error) {
	cast, ok := in.(proto.Message)
	if !ok {
		return []byte{}, ErrNotProtoMessage
	}

	out, err := proto.Marshal(cast)
	return out, err
}

func (p ProtoEncoder) Unmarshal(data []byte, obj interface{}) error {
	cast, ok := obj.(proto.Message)
	if !ok {
		return ErrNotProtoMessage
	}
	err := proto.Unmarshal(data, cast)
	return err
}

// Json encoder/dedboder
type JsonEncoder struct{}

func (p JsonEncoder) Marshal(in interface{}) ([]byte, error) {
	out, err := json.Marshal(in)
	return out, err
}

func (p JsonEncoder) Unmarshal(data []byte, obj interface{}) error {
	err := json.Unmarshal(data, obj)
	return err
}
