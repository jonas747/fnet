package fnet

import (
	"errors"
	"github.com/golang/protobuf/proto"
)

type Encoder interface {
	Marshal(in interface{}) ([]byte, error)
	Unmarshal(data []byte, obj interface{}) error
}

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
