package connectrpc

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// MaxDeliveryMessages caps what decoding a connector's answer may allocate
// (D-184): its messages, nested ones included, and the elements of its
// repeated fields together, counted on the wire before anything is decoded.
// An empty message costs two bytes on the wire and about 90 bytes once
// decoded, so the byte cap alone would let a connector make the core hold
// gigabytes. A backfill of 200,000 daily prices (three messages each) stays
// below it.
const MaxDeliveryMessages = 640_000

// maxWireDepth caps how deep the count follows nested messages. The
// connector contract nests a few levels; protobuf's own limit is 10,000.
const maxWireDepth = 32

// Sizes on the wire: a fixed32 and a fixed64 value, and the bit of a
// varint's byte that says another byte follows.
const (
	fixed32Size = 4
	fixed64Size = 8
	varintMore  = 0x80
)

// errMalformedWire reports an answer that isn't valid protobuf wire format.
var errMalformedWire = errors.New("malformed protobuf")

// boundedCodec is connect's binary protobuf codec, under the same name, that
// counts a message's elements on the wire, without allocating, and refuses
// one of more than MaxDeliveryMessages with ErrDeliveryTooLarge before
// decoding it.
type boundedCodec struct{}

// Name returns "proto", the name of the codec it replaces.
func (boundedCodec) Name() string { return "proto" }

// Marshal encodes message, which must be a proto.Message.
func (boundedCodec) Marshal(message any) ([]byte, error) {
	m, ok := message.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("%w: %T is not a protobuf message", errMalformedWire, message)
	}
	b, err := proto.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return b, nil
}

// Unmarshal decodes data into message once its elements are counted
// within MaxDeliveryMessages, the message itself included.
func (boundedCodec) Unmarshal(data []byte, message any) error {
	m, ok := message.(proto.Message)
	if !ok {
		return fmt.Errorf("%w: %T is not a protobuf message", errMalformedWire, message)
	}
	if _, err := countElements(data, m.ProtoReflect().Descriptor(), MaxDeliveryMessages-1, 0); err != nil {
		return err
	}
	if err := proto.Unmarshal(data, m); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	return nil
}

// countElements counts the messages nested in b, a message of type md, and
// the elements of its repeated fields, and returns how many of budget are
// left. It fails with ErrDeliveryTooLarge as soon as the count passes budget
// and with errMalformedWire on bytes that aren't wire format. Unknown fields
// cost only their bytes once decoded, so they are skipped, as is anything
// past maxWireDepth, which fails.
func countElements(b []byte, md protoreflect.MessageDescriptor, budget, depth int) (int, error) {
	if depth > maxWireDepth {
		return 0, fmt.Errorf("%w: nested more than %d deep", errMalformedWire, maxWireDepth)
	}
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return 0, fmt.Errorf("%w: %w", errMalformedWire, protowire.ParseError(n))
		}
		b = b[n:]
		size := protowire.ConsumeFieldValue(num, typ, b)
		if size < 0 {
			return 0, fmt.Errorf("%w: %w", errMalformedWire, protowire.ParseError(size))
		}
		value := b[:size]
		b = b[size:]
		fd := md.Fields().ByNumber(num)
		if fd == nil {
			continue
		}
		var err error
		if budget, err = countField(fd, typ, value, budget, depth); err != nil {
			return 0, err
		}
	}
	return budget, nil
}

// countField counts one occurrence of field fd, value its bytes after the
// tag, against budget, and returns what is left.
func countField(fd protoreflect.FieldDescriptor, typ protowire.Type, value []byte, budget, depth int) (int, error) {
	if fd.Message() != nil && typ == protowire.BytesType {
		v, _ := protowire.ConsumeBytes(value)
		budget--
		if budget < 0 {
			return 0, tooMany()
		}
		return countElements(v, fd.Message(), budget, depth+1)
	}
	if budget -= scalarElements(fd, typ, value); budget < 0 {
		return 0, tooMany()
	}
	return budget, nil
}

// scalarElements is how many elements one occurrence of a field that isn't
// a length-delimited message adds: those of a packed list, one for any
// other list element or a group, none for a singular scalar.
func scalarElements(fd protoreflect.FieldDescriptor, typ protowire.Type, value []byte) int {
	k := fd.Kind()
	if fd.IsList() && typ == protowire.BytesType && k != protoreflect.StringKind && k != protoreflect.BytesKind {
		v, _ := protowire.ConsumeBytes(value)
		return packedElements(k, v)
	}
	if fd.IsList() || fd.Message() != nil {
		return 1
	}
	return 0
}

// packedElements is an upper bound on the elements of a packed repeated
// field of kind k: one per varint's last byte, or per fixed-size value.
func packedElements(k protoreflect.Kind, v []byte) int {
	switch protowireType(k) {
	case protowire.Fixed32Type:
		return len(v) / fixed32Size
	case protowire.Fixed64Type:
		return len(v) / fixed64Size
	case protowire.VarintType, protowire.BytesType, protowire.StartGroupType, protowire.EndGroupType:
	}
	n := 0
	for _, c := range v {
		if c&varintMore == 0 {
			n++
		}
	}
	return n
}

// protowireType is the wire type of a scalar of kind k.
func protowireType(k protoreflect.Kind) protowire.Type {
	switch k {
	case protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind, protoreflect.FloatKind:
		return protowire.Fixed32Type
	case protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind, protoreflect.DoubleKind:
		return protowire.Fixed64Type
	case protoreflect.BoolKind, protoreflect.EnumKind, protoreflect.Int32Kind, protoreflect.Sint32Kind,
		protoreflect.Uint32Kind, protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
		protoreflect.StringKind, protoreflect.BytesKind, protoreflect.MessageKind, protoreflect.GroupKind:
	}
	return protowire.VarintType
}

// tooMany is the error for an answer over MaxDeliveryMessages.
func tooMany() error {
	return fmt.Errorf("%w: more than %d messages and repeated elements", ErrDeliveryTooLarge, MaxDeliveryMessages)
}
