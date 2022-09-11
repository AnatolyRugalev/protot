package protoxtest

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/AnatolyRugalev/protox"
)

func NewAssertions(t assert.TestingT) *Assertions {
	return &Assertions{
		assert: assert.New(t),
	}
}

type Assertions struct {
	assert *assert.Assertions
}

func (a Assertions) Equal(expected, actual proto.Message, msgAndArgs ...any) bool {
	if proto.Equal(expected, actual) {
		return true
	}
	if !a.assert.Equal(expected.ProtoReflect().Descriptor().FullName(), actual.ProtoReflect().Descriptor().FullName(), "message types are not equal") {
		return false
	}
	diff, err := GetJSONDiff(expected, actual)
	if !a.assert.NoError(err, "error formatting protobuf message diff") {
		return false
	}
	return a.assert.Fail("Not equal:\n"+diff.Diff.Render(), msgAndArgs...)
}

func (a Assertions) EqualWire(msgA, msgB proto.Message, msgAndArgs ...any) bool {
	marshal := proto.MarshalOptions{
		Deterministic: true,
	}
	unmarshal := proto.UnmarshalOptions{
		DiscardUnknown: true,
	}

	var oneofs map[protoreflect.Name]protoreflect.OneofDescriptor
	var oneofsA map[protoreflect.Name]protoreflect.OneofDescriptor
	var oneofsB map[protoreflect.Name]protoreflect.OneofDescriptor
	fuzzer := NewFuzzer(time.Now().UnixNano(), FuzzWithOneofHandler(func(_ *Fuzzer, _ protoreflect.Message, o map[protoreflect.Name]protoreflect.OneofDescriptor) {
		oneofs = o
	}))

	var err error
	aDesc := msgA.ProtoReflect().Descriptor()
	bDesc := msgB.ProtoReflect().Descriptor()

	actualA := proto.Clone(msgA)
	fuzzer.Message(msgA)
	oneofsA = oneofs
	err = protox.ConvertWithOptions(msgA, msgB, marshal, unmarshal, false)
	if !a.assert.NoError(err, "error convering %s to %s", aDesc.FullName(), bDesc.FullName()) {
		return false
	}
	err = protox.ConvertWithOptions(msgB, actualA, marshal, unmarshal, false)
	if !a.assert.NoError(err, "error convering %s to %s", bDesc.FullName(), aDesc.FullName()) {
		return false
	}
	if !a.Equal(msgA, actualA, msgAndArgs...) {
		return false
	}

	actualB := proto.Clone(msgB)
	fuzzer.Message(msgB)
	oneofsB = oneofs
	err = protox.ConvertWithOptions(msgB, msgA, marshal, unmarshal, false)
	if !a.assert.NoError(err, "error convering %s to %s", bDesc.FullName(), aDesc.FullName()) {
		return false
	}
	err = protox.ConvertWithOptions(msgA, actualB, marshal, unmarshal, false)
	if !a.assert.NoError(err, "error convering %s to %s", aDesc.FullName(), bDesc.FullName()) {
		return false
	}
	if !a.Equal(msgA, actualA) {
		a.assert.Fail("Message wire formats are not equal", msgAndArgs...)
		return false
	}
	if !a.EqualWireOneofs(oneofsA, oneofsB) {
		a.assert.Fail("Message wire formats are not equal", msgAndArgs...)
		return false
	}
	return true
}

func (a Assertions) EqualWireOneofs(oA, oB map[protoreflect.Name]protoreflect.OneofDescriptor) bool {
	if oA == nil && oB == nil {
		return true
	}
	oneofsA := map[string]protoreflect.OneofDescriptor{}
	for _, oneof := range oA {
		fields := oneof.Fields()
		hashParts := make([]string, 0, fields.Len())
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			hashParts = append(hashParts, strconv.Itoa(int(field.Number())))
		}
		hash := strings.Join(hashParts, ",")
		oneofsA[hash] = oneof
	}

	oneofsB := map[string]protoreflect.OneofDescriptor{}
	for _, oneof := range oB {
		fields := oneof.Fields()
		hashParts := make([]string, 0, fields.Len())
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			hashParts = append(hashParts, strconv.Itoa(int(field.Number())))
		}
		hash := strings.Join(hashParts, ",")
		oneofsB[hash] = oneof
	}

	equal := true
	visited := make(map[string]struct{})
	for hash, oneofA := range oneofsA {
		oneofB, ok := oneofsB[hash]
		visited[hash] = struct{}{}
		if !ok {
			equal = false
			a.assert.Fail(fmt.Sprintf("oneof %s does not have matching oneof", oneofA.FullName()))
			continue
		}
		fieldsA := oneofA.Fields()
		for i := 0; i < fieldsA.Len(); i++ {
			fieldA := fieldsA.Get(i)
			fieldB := oneofB.Fields().ByNumber(fieldA.Number())
			if fieldA.Kind() != fieldB.Kind() {
				equal = false
				a.assert.Fail(fmt.Sprintf("type of field %s (%d) does not match field %s", fieldA.FullName(), fieldA.Number(), fieldB.FullName()))
				continue
			}
			if fieldA.IsMap() {
				if fieldA.MapKey().Kind() != fieldB.MapKey().Kind() {
					equal = false
					a.assert.Fail(fmt.Sprintf("map key type of field %s (%d) does not match field %s", fieldA.FullName(), fieldA.Number(), fieldB.FullName()))
					continue
				}
			}
			if fieldA.Kind() == protoreflect.MessageKind {
				messageA := dynamicpb.NewMessage(fieldA.Message())
				messageB := dynamicpb.NewMessage(fieldB.Message())
				a.EqualWire(messageA, messageB)
			}
		}
	}

	for hash, oneofB := range oneofsB {
		if _, ok := visited[hash]; ok {
			continue
		}
		equal = false
		a.assert.Fail(fmt.Sprintf("oneof %s is not present in message A", oneofB.FullName()))
	}

	return equal
}

func (a Assertions) Subset(superset, subset proto.Message, msgAndArgs ...any) bool {
	if proto.Equal(superset, subset) {
		return true
	}
	ok, diff, err := IsSubset(superset, subset)
	if !a.assert.NoError(err, "error evaluating subset") {
		return false
	}
	if ok {
		return true
	}
	return a.assert.Fail("Not a subset:\n"+diff.Diff.Render(), msgAndArgs...)
}

var protoMsgType = reflect.TypeOf(new(proto.Message))

func (a Assertions) SliceElemFunc(slice any, fn func(m proto.Message) bool) bool {
	typ := reflect.TypeOf(slice)
	if !a.assert.Equal(typ.Kind(), reflect.Slice, "slice type expected, got %T", slice) {
		return false
	}
	if !a.assert.True(typ.Elem().AssignableTo(protoMsgType), "slice of proto.Message expected, got %T", slice) {
		return false
	}
	val := reflect.ValueOf(slice)
	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i).Interface().(proto.Message)
		if !fn(elem) {
			a.assert.Fail("SliceElemFunc returned false")
			return false
		}
	}
	return true
}

func (a Assertions) SliceElemSubset(slice any, subset proto.Message, target *proto.Message) bool {
	var found proto.Message
	a.SliceElemFunc(slice, func(m proto.Message) bool {
		if found != nil {
			return true
		}
		ok, _, err := IsSubset(m, subset)
		if !a.assert.NoError(err, "error evaluating subset") {
			return false
		}
		if ok {
			found = m
		}
		return true
	})
	if found != nil {
		if target != nil {
			*target = found
		}
		return true
	}
	return a.assert.Fail("A subset was not found in provided slice")
}

func IsSubset(superset, subset proto.Message) (bool, *JSONDiff, error) {
	jsonDiff, err := GetJSONDiff(superset, subset)
	if err != nil {
		return false, nil, err
	}
	ok := true
	for _, op := range jsonDiff.Patch {
		if op.Op != "remove" && op.Op != "test" {
			ok = false
			break
		}
	}
	return ok, jsonDiff, nil
}

//
//func ContainsArrayElementSubset[R proto.Message](t *testing.T, array []R, expected R) R {
//	match := ArrayElementSubset(t, array, expected)
//	require.NotNil(t, match, "array does not contain expected PB message subset")
//	return match
//}
//
//func NotContainsArrayElementSubset[R proto.Message](t *testing.T, array []R, expected R) {
//	match := ArrayElementSubset(t, array, expected)
//	require.Nil(t, match, "array contains unexpected PB message subset")
//}
