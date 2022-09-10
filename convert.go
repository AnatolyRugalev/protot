package protox

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

type VTMessage interface {
	proto.Message
	Reset()
	MarshalVT() ([]byte, error)
	UnmarshalVT(data []byte) error
}

// ConvertVT is a VTProto version of Convert
func ConvertVT(source VTMessage, target VTMessage) error {
	bytes, err := source.MarshalVT()
	if err != nil {
		return fmt.Errorf("error marshalling %T: %w", source, err)
	}
	target.Reset()
	err = target.UnmarshalVT(bytes)
	if err != nil {
		return fmt.Errorf("error unmarshalling into %T: %w", target, err)
	}
	return nil
}

// Convert is a shorthand for ConvertWithOptions
func Convert(source proto.Message, target proto.Message) error {
	return ConvertWithOptions(source, target, proto.MarshalOptions{}, proto.UnmarshalOptions{})
}

// ConvertWithOptions converts source PB message to target message. Only works with protobuf wire-compatible messages.
// Provided message wire formats should be in sync: field structure, field numbers and field types.
func ConvertWithOptions(source proto.Message, target proto.Message, marshal proto.MarshalOptions, unmarshal proto.UnmarshalOptions) error {
	bytes, err := marshal.Marshal(source)
	if err != nil {
		return fmt.Errorf("error marshalling %T: %w", source, err)
	}
	err = unmarshal.Unmarshal(bytes, target)
	if err != nil {
		return fmt.Errorf("error unmarshalling into %T: %w", target, err)
	}
	return nil
}
