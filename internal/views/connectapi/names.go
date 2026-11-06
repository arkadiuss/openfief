package connectapi

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/arkadiuss/openfief/internal/views/app"
)

// The messages that name an instrument by its ISIN, whose name field the
// read API fills in with its display name (D-197).
const (
	instrumentMessage protoreflect.FullName = "openfief.views.v1.Instrument"
	detailMessage     protoreflect.FullName = "openfief.views.v1.InstrumentDetail"
)

// namesInterceptor fills in, on every answer, the display name of each
// instrument it names (D-197): Instrument.name and InstrumentDetail.name,
// read from the views once per answer that names any. An instrument with
// no name keeps an empty one.
type namesInterceptor struct {
	views *app.Views
	log   *slog.Logger
}

// WrapUnary names the instruments of next's answer.
func (n *namesInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		resp, err := next(ctx, req)
		if err != nil || resp == nil {
			return resp, err
		}
		msg, ok := resp.Any().(proto.Message)
		if !ok {
			return resp, nil
		}
		named := namedInstruments(msg.ProtoReflect(), nil)
		if len(named) == 0 {
			return resp, nil
		}
		names, err := n.views.InstrumentNames(ctx)
		if err != nil {
			return nil, toConnectError(ctx, n.log, err)
		}
		fillNames(named, names)
		return resp, nil
	}
}

// WrapStreamingClient passes a client stream through: the read API has
// none.
func (*namesInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler passes a handler stream through: the read API has
// none.
func (*namesInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

// namedInstruments appends to out every Instrument and InstrumentDetail
// message within m, m itself included, and returns out.
func namedInstruments(m protoreflect.Message, out []protoreflect.Message) []protoreflect.Message {
	if n := m.Descriptor().FullName(); n == instrumentMessage || n == detailMessage {
		out = append(out, m)
	}
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if fd.Message() == nil || fd.IsMap() {
			return true
		}
		if fd.IsList() {
			l := v.List()
			for i := range l.Len() {
				out = namedInstruments(l.Get(i).Message(), out)
			}
			return true
		}
		out = namedInstruments(v.Message(), out)
		return true
	})
	return out
}

// fillNames sets each of named's name field to the display name of its
// ISIN in names, when it has one.
func fillNames(named []protoreflect.Message, names map[string]string) {
	for _, m := range named {
		fields := m.Descriptor().Fields()
		isin := m.Get(fields.ByName("isin")).String()
		if name, ok := names[isin]; ok {
			m.Set(fields.ByName("name"), protoreflect.ValueOfString(name))
		}
	}
}
