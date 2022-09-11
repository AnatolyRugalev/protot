package protox

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

type vtMarshal interface {
	proto.Message
	MarshalVT() ([]byte, error)
}

type vtUnmarshal interface {
	proto.Message
	Reset()
	UnmarshalVT(data []byte) error
}

// Convert is a shortcut for ConvertWithOptions
func Convert(source proto.Message, target proto.Message) error {
	return ConvertWithOptions(source, target, proto.MarshalOptions{}, proto.UnmarshalOptions{}, true)
}

// ConvertWithOptions converts source PB message to target message. Only works with protobuf wire-compatible messages.
// Provided message wire formats should be in sync: field structure, field numbers and field types.
func ConvertWithOptions(source proto.Message, target proto.Message, marshal proto.MarshalOptions, unmarshal proto.UnmarshalOptions, tryVT bool) error {
	var (
		bytes []byte
		err   error
	)
	vtOk := false
	if tryVT {
		if vt, ok := source.(vtMarshal); ok {
			vtOk = true
			bytes, err = vt.MarshalVT()
		}
	}
	if !vtOk {
		bytes, err = marshal.Marshal(source)
	}
	if err != nil {
		return fmt.Errorf("error marshalling %T: %w", source, err)
	}

	vtOk = false
	if tryVT {
		if vt, ok := target.(vtUnmarshal); ok {
			vtOk = true
			err = vt.UnmarshalVT(bytes)
		}
	}
	if !vtOk {
		err = unmarshal.Unmarshal(bytes, target)
	}
	if err != nil {
		return fmt.Errorf("error unmarshalling into %T: %w", target, err)
	}
	return nil
}
