package protoxtest

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

type TextDiff struct {
	Expected []byte
	Actual   []byte
}

func GetTextDiff(expected, actual proto.Message) (*TextDiff, error) {
	opts := prototext.MarshalOptions{
		Multiline:    true,
		Indent:       "\t",
		AllowPartial: true,
		EmitUnknown:  true,
	}
	expectedText, err := opts.Marshal(expected)
	if err != nil {
		return nil, fmt.Errorf("error marshalling expected proto %T into JSON: %w", expected, err)
	}
	actualText, err := opts.Marshal(actual)
	if err != nil {
		return nil, fmt.Errorf("error marshalling actual proto %T into JSON: %w", actual, err)
	}
	return &TextDiff{
		Expected: expectedText,
		Actual:   actualText,
	}, nil
}

func FormatTextDiff(expected, actual proto.Message) (string, error) {
	diff, err := GetTextDiff(expected, actual)
	if err != nil {
		return "", err
	}
	return cmp.Diff(diff.Expected, diff.Actual), nil
}
