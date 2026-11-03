package connectrpc

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/testing/protocmp"

	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// TestBoundedCodecCounts: the codec counts every message, nested ones
// included, and every repeated element on the wire, and refuses an answer
// of more than MaxDeliveryMessages with ErrDeliveryTooLarge before decoding
// it; anything within the cap decodes as protobuf does (D-184).
func TestBoundedCodecCounts(t *testing.T) {
	t.Parallel()
	posting := wireField(nil, 7, nil)
	// A SyncResponse, its Delivery and one Transaction are three messages.
	tests := []struct {
		name     string
		postings int
		wantErr  error
	}{
		{name: "one posting", postings: 1},
		{name: "exactly the cap", postings: MaxDeliveryMessages - 3},
		{name: "one over the cap", postings: MaxDeliveryMessages - 2, wantErr: ErrDeliveryTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tx := wireField(nil, 3, bytes.Repeat(posting, tt.postings))
			data := wireField(nil, 1, append(wireField(nil, 1, []byte("b-1")), tx...))
			var got connectorv1.SyncResponse
			err := boundedCodec{}.Unmarshal(data, &got)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Unmarshal error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if got.GetDelivery() != nil {
					t.Error("a refused answer was decoded")
				}
				return
			}
			if n := len(got.GetDelivery().GetTransactions()[0].GetPostings()); n != tt.postings {
				t.Errorf("decoded %d postings, want %d", n, tt.postings)
			}
		})
	}
}

// TestBoundedCodecRoundTrip: what the codec marshals it unmarshals
// unchanged, unknown fields are skipped by the count, and malformed wire
// bytes or a value that isn't a protobuf message are refused.
func TestBoundedCodecRoundTrip(t *testing.T) {
	t.Parallel()
	want := &connectorv1.SyncResponse{Watermark: "w-1", Delivery: &connectorv1.Delivery{
		BatchId: "b-1",
		Prices: []*connectorv1.Price{{
			Instrument: &connectorv1.Instrument{Isin: "PLTEST000010", Symbol: "TST"},
			Date:       "2024-01-02", Close: &connectorv1.Money{Amount: "1.00", Currency: "PLN"},
		}},
	}}
	data, err := boundedCodec{}.Marshal(want)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// An unknown field: a large number with a nested-looking value.
	data = wireField(data, 99, bytes.Repeat(wireField(nil, 7, nil), 10))
	var got connectorv1.SyncResponse
	if err := (boundedCodec{}).Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if diff := cmp.Diff(want, &got, protocmp.Transform(), protocmp.IgnoreUnknown()); diff != "" {
		t.Errorf("round trip (-want +got):\n%s", diff)
	}
	if err := (boundedCodec{}).Unmarshal([]byte{0x0a, 0x05, 0x01}, &got); !errors.Is(err, errMalformedWire) {
		t.Errorf("Unmarshal of truncated bytes = %v, want %v", err, errMalformedWire)
	}
	if err := (boundedCodec{}).Unmarshal(data, "not a message"); !errors.Is(err, errMalformedWire) {
		t.Errorf("Unmarshal into a string = %v, want %v", err, errMalformedWire)
	}
	if _, err := (boundedCodec{}).Marshal("not a message"); !errors.Is(err, errMalformedWire) {
		t.Errorf("Marshal of a string = %v, want %v", err, errMalformedWire)
	}
	if name := (boundedCodec{}).Name(); name != "proto" {
		t.Errorf("Name = %q, want proto", name)
	}
}

// TestCountElementsRepeatedScalars: a packed repeated field counts each
// element, and an unpacked repeated string counts each occurrence, so a
// request's lists cost what they decode to.
func TestCountElementsRepeatedScalars(t *testing.T) {
	t.Parallel()
	md := (&connectorv1.SyncRequest{}).ProtoReflect().Descriptor()
	indices := bytes.Repeat(wireField(nil, 5, nil), 4)
	left, err := countElements(indices, md, 10, 0)
	if err != nil || left != 6 {
		t.Errorf("countElements of 4 indices = %d, %v; want 6 left", left, err)
	}
	if _, err := countElements(indices, md, 3, 0); !errors.Is(err, ErrDeliveryTooLarge) {
		t.Errorf("countElements over budget = %v, want %v", err, ErrDeliveryTooLarge)
	}
	tests := []struct {
		name string
		kind protoreflect.Kind
		v    []byte
		want int
	}{
		{name: "varints", kind: protoreflect.Int64Kind, v: []byte{0x01, 0x81, 0x01, 0x7f}, want: 3},
		{name: "fixed32", kind: protoreflect.Fixed32Kind, v: make([]byte, 12), want: 3},
		{name: "fixed64", kind: protoreflect.Fixed64Kind, v: make([]byte, 16), want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := packedElements(tt.kind, tt.v); got != tt.want {
				t.Errorf("packedElements = %d, want %d", got, tt.want)
			}
		})
	}
}

// wireField appends field num of wire type bytes holding v to b.
func wireField(b []byte, num protowire.Number, v []byte) []byte {
	return protowire.AppendBytes(protowire.AppendTag(b, num, protowire.BytesType), v)
}
