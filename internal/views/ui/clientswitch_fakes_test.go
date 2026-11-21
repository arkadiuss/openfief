package ui

import (
	"context"
	"sync/atomic"

	"connectrpc.com/connect"

	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
)

// fakeClientEvents is the fake core's switch for the UI's client events (D-227):
// on until the owner turns it off, read without a session, changed only by the
// owner, and counting how often it was asked.
type fakeClientEvents struct {
	off   atomic.Bool
	reads atomic.Int32
	sets  atomic.Int32
	down  atomic.Bool
}

func (f *fakeSettings) GetClientEvents(context.Context, *connect.Request[valuationv1.GetClientEventsRequest]) (*connect.Response[valuationv1.GetClientEventsResponse], error) {
	f.events.reads.Add(1)
	if f.events.down.Load() {
		return nil, codeErr(connect.CodeUnavailable, nil)
	}
	return connect.NewResponse(&valuationv1.GetClientEventsResponse{Enabled: !f.events.off.Load()}), nil
}

func (f *fakeSettings) SetClientEvents(_ context.Context, r *connect.Request[valuationv1.SetClientEventsRequest]) (*connect.Response[valuationv1.SetClientEventsResponse], error) {
	if r.Msg.GetSessionToken() != ownerSession {
		return nil, codeErr(connect.CodePermissionDenied, errNotOwnerFake)
	}
	f.events.sets.Add(1)
	f.events.off.Store(!r.Msg.GetEnabled())
	return connect.NewResponse(&valuationv1.SetClientEventsResponse{}), nil
}
