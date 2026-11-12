package main

import (
	"context"
	"io"
	"log/slog"
	"slices"
	"sync"
	"testing"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"google.golang.org/grpc"

	"github.com/arkadiuss/openfief/internal/identity/spicedb"
)

// writeLog is a spicedb.Permissions that holds nothing and records the
// relationships written, as "operation type:id#relation@type:id".
type writeLog struct {
	mu     sync.Mutex
	writes []string
}

var _ spicedb.Permissions = (*writeLog)(nil)

func (*writeLog) CheckPermission(context.Context, *v1.CheckPermissionRequest, ...grpc.CallOption) (*v1.CheckPermissionResponse, error) {
	return &v1.CheckPermissionResponse{}, nil
}

func (*writeLog) LookupResources(context.Context, *v1.LookupResourcesRequest, ...grpc.CallOption) (grpc.ServerStreamingClient[v1.LookupResourcesResponse], error) {
	return emptyStream[v1.LookupResourcesResponse]{}, nil
}

func (*writeLog) ReadRelationships(context.Context, *v1.ReadRelationshipsRequest, ...grpc.CallOption) (grpc.ServerStreamingClient[v1.ReadRelationshipsResponse], error) {
	return emptyStream[v1.ReadRelationshipsResponse]{}, nil
}

func (w *writeLog) WriteRelationships(_ context.Context, in *v1.WriteRelationshipsRequest, _ ...grpc.CallOption) (*v1.WriteRelationshipsResponse, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, u := range in.GetUpdates() {
		r := u.GetRelationship()
		w.writes = append(w.writes, u.GetOperation().String()+" "+r.GetResource().GetObjectType()+":"+r.GetResource().GetObjectId()+
			"#"+r.GetRelation()+"@"+r.GetSubject().GetObject().GetObjectType()+":"+r.GetSubject().GetObject().GetObjectId())
	}
	return &v1.WriteRelationshipsResponse{}, nil
}

// emptyStream is a stream with nothing in it.
type emptyStream[T any] struct{ grpc.ClientStream }

func (emptyStream[T]) Recv() (*T, error) { return nil, io.EOF }

// TestPlacementSyncPassesIDsOn checks that the registry's placements reach
// the syncer as the account's household and sleeve, and a removal as the
// end of the sleeve relationship alone.
func TestPlacementSyncPassesIDsOn(t *testing.T) {
	t.Parallel()
	w := &writeLog{}
	p := placementSync{spicedb.NewSyncer(w)}
	if err := p.AccountPlaced(t.Context(), "brokerage", "core"); err != nil {
		t.Fatalf("AccountPlaced: %v", err)
	}
	want := []string{
		"OPERATION_TOUCH account:brokerage#household@household:home",
		"OPERATION_TOUCH account:brokerage#sleeve@sleeve:core",
	}
	if !slices.Equal(w.writes, want) {
		t.Errorf("writes = %v, want %v", w.writes, want)
	}
	if err := p.AccountRemoved(t.Context(), "brokerage"); err != nil {
		t.Fatalf("AccountRemoved: %v", err)
	}
	if len(w.writes) != len(want) {
		t.Errorf("a removal wrote %v with nothing held", w.writes[len(want):])
	}
}

// TestReconcileAccessWritesTheOwner checks that the start's reconcile writes
// the household's owner and each account of the loader, and does nothing
// without a syncer, as with the policy.
func TestReconcileAccessWritesTheOwner(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.DiscardHandler)
	load := func(context.Context) (spicedb.Desired, error) {
		return spicedb.Desired{Accounts: []string{"bonds"}, Sleeves: map[string]string{"bonds": "bonds"}}, nil
	}
	reconcileAccess(t.Context(), nil, load, log) // the policy: nothing to do, nothing to fail
	w := &writeLog{}
	reconcileAccess(t.Context(), spicedb.NewSyncer(w), load, log)
	want := []string{
		"OPERATION_TOUCH household:home#owner@person:owner",
		"OPERATION_TOUCH account:bonds#household@household:home",
		"OPERATION_TOUCH account:bonds#sleeve@sleeve:bonds",
	}
	if !slices.Equal(w.writes, want) {
		t.Errorf("writes = %v, want %v", w.writes, want)
	}
}
