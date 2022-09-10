package protoxtest

import (
	"encoding/json"
	"fmt"

	jd "github.com/josephburnett/jd/lib"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type JSONOperation struct {
	Op    string          `json:"op"`
	Path  string          `json:"path"`
	Value json.RawMessage `json:"value"`
}

type JSONDiff struct {
	Expected jd.JsonNode
	Actual   jd.JsonNode
	Diff     jd.Diff
	Patch    []JSONOperation
}

func GetJSONDiff(expected, actual proto.Message) (*JSONDiff, error) {
	opts := protojson.MarshalOptions{
		Multiline:      true,
		Indent:         "\t",
		AllowPartial:   true,
		UseProtoNames:  true,
		UseEnumNumbers: false,
	}
	expectedJSON, err := opts.Marshal(expected)
	if err != nil {
		return nil, fmt.Errorf("error marshalling expected proto %T into JSON: %w", expected, err)
	}
	actualJSON, err := opts.Marshal(actual)
	if err != nil {
		return nil, fmt.Errorf("error marshalling actual proto %T into JSON: %w", actual, err)
	}
	jdActual, err := jd.ReadJsonString(string(actualJSON))
	if err != nil {
		return nil, err
	}
	jdExpected, err := jd.ReadJsonString(string(expectedJSON))
	if err != nil {
		return nil, err
	}
	diff := jdExpected.Diff(jdActual)
	patch, err := diff.RenderPatch()
	if err != nil {
		return nil, err
	}
	var ops []JSONOperation
	err = json.Unmarshal([]byte(patch), &ops)
	if err != nil {
		return nil, err
	}
	return &JSONDiff{
		Expected: jdActual,
		Actual:   jdExpected,
		Patch:    ops,
		Diff:     diff,
	}, nil
}
