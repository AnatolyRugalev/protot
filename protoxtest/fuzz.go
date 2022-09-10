package protoxtest

import (
	"fmt"
	"math/rand"
	"time"

	fuzz "github.com/google/gofuzz"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type OneofHandlerFunc func(f *Fuzzer, msgReflect protoreflect.Message, oneofs map[protoreflect.Name]protoreflect.OneofDescriptor)

type Fuzzer struct {
	fuzz         *fuzz.Fuzzer
	rand         *rand.Rand
	oneofHandler OneofHandlerFunc
}

func FuzzWithOneofHandler(fn OneofHandlerFunc) FuzzOption {
	return func(f *Fuzzer) {
		f.oneofHandler = fn
	}
}

func NewFuzzer(seed int64, opts ...FuzzOption) *Fuzzer {
	source := rand.NewSource(seed)
	fuzzer := &Fuzzer{
		fuzz: fuzz.New().NilChance(0).NumElements(1, 1).RandSource(source),
		rand: rand.New(source),
		oneofHandler: func(f *Fuzzer, msgReflect protoreflect.Message, oneofs map[protoreflect.Name]protoreflect.OneofDescriptor) {
			for _, oneof := range oneofs {
				oFields := oneof.Fields()
				field := oFields.Get(f.rand.Intn(oFields.Len()))
				f.Field(msgReflect, field)
			}
		},
	}
	for _, o := range opts {
		o(fuzzer)
	}
	return fuzzer
}

type FuzzOption func(f *Fuzzer)

func Fuzz(m proto.Message, opts ...FuzzOption) {
	FuzzWithSeed(m, time.Now().UnixNano(), opts...)
}

func FuzzWithSeed(m proto.Message, seed int64, opts ...FuzzOption) {
	NewFuzzer(seed, opts...).Message(m)
}

func (f *Fuzzer) Message(m proto.Message) {
	msgReflect := m.ProtoReflect()
	fields := msgReflect.Descriptor().Fields()
	oneofs := map[protoreflect.Name]protoreflect.OneofDescriptor{}
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		oneof := field.ContainingOneof()
		if oneof != nil {
			oneofs[field.Name()] = oneof
			continue
		}
		f.Field(msgReflect, field)
	}
	if len(oneofs) > 0 {
		f.oneofHandler(f, msgReflect, oneofs)
	}
}

func (f *Fuzzer) Field(msgReflect protoreflect.Message, field protoreflect.FieldDescriptor) {
	protoValue := msgReflect.Get(field)
	switch field.Kind() {
	case protoreflect.EnumKind:
		values := field.Enum().Values()
		value := values.Get(f.rand.Intn(values.Len()))
		msgReflect.Set(field, protoreflect.ValueOf(value))
	case protoreflect.MessageKind:
		message := protoValue.Message().New()
		f.Message(message.Interface())
		msgReflect.Set(field, protoreflect.ValueOf(message))
	case
		protoreflect.Int64Kind,
		protoreflect.Int32Kind,
		protoreflect.Uint64Kind,
		protoreflect.Uint32Kind:
		val := protoValue.Int()
		f.fuzz.Fuzz(&val)
		msgReflect.Set(field, protoreflect.ValueOf(val))
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		val := protoValue.Float()
		f.fuzz.Fuzz(&val)
		msgReflect.Set(field, protoreflect.ValueOf(val))
	case protoreflect.BoolKind:
		val := protoValue.Bool()
		f.fuzz.Fuzz(&val)
		msgReflect.Set(field, protoreflect.ValueOf(val))
	case protoreflect.StringKind:
		val := protoValue.String()
		f.fuzz.Fuzz(&val)
		msgReflect.Set(field, protoreflect.ValueOf(val))
	case protoreflect.BytesKind:
		val := protoValue.Bytes()
		f.fuzz.Fuzz(&val)
		msgReflect.Set(field, protoreflect.ValueOf(val))
	default:
		panic(fmt.Sprintf("unsupported proto kind: %s", field.Kind().String()))
	}
}
