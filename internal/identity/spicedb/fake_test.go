package spicedb_test

import (
	"context"
	"fmt"
	"io"
	"slices"
	"sync"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"google.golang.org/grpc"

	"github.com/arkadiuss/openfief/internal/identity/spicedb"
)

// stream is a grpc.ServerStreamingClient that returns items, then err, or
// io.EOF when err is nil.
type stream[T any] struct {
	grpc.ClientStream

	items []*T
	err   error
}

func (s *stream[T]) Recv() (*T, error) {
	if len(s.items) == 0 {
		if s.err != nil {
			return nil, s.err
		}
		return nil, io.EOF
	}
	it := s.items[0]
	s.items = s.items[1:]
	return it, nil
}

// fake is spicedb.Permissions over relationships kept in memory. It answers
// checks and lookups from the scripted fields, which a test sets, and reads
// and writes relationships as SpiceDB does, without the schema's rules.
type fake struct {
	mu sync.Mutex

	// check answers CheckPermission, unless err is set.
	check v1.CheckPermissionResponse_Permissionship
	// lookup is what LookupResources streams; lookupErr ends the stream
	// after it.
	lookup    []*v1.LookupResourcesResponse
	lookupErr error
	// err fails every call; block makes every call wait for its context.
	err   error
	block bool

	// tuples are the relationships held, in zed's notation.
	tuples []spicedb.Tuple
	// checks, lookups and writes record the requests; reads count the
	// ReadRelationships calls.
	checks  []*v1.CheckPermissionRequest
	lookups []*v1.LookupResourcesRequest
	writes  []*v1.WriteRelationshipsRequest
	reads   int
}

var _ spicedb.Permissions = (*fake)(nil)

func (f *fake) CheckPermission(ctx context.Context, in *v1.CheckPermissionRequest, _ ...grpc.CallOption) (*v1.CheckPermissionResponse, error) {
	if err := f.wait(ctx); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.checks = append(f.checks, in)
	return &v1.CheckPermissionResponse{Permissionship: f.check}, nil
}

func (f *fake) LookupResources(ctx context.Context, in *v1.LookupResourcesRequest, _ ...grpc.CallOption) (grpc.ServerStreamingClient[v1.LookupResourcesResponse], error) {
	if err := f.wait(ctx); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lookups = append(f.lookups, in)
	return &stream[v1.LookupResourcesResponse]{items: slices.Clone(f.lookup), err: f.lookupErr}, nil
}

func (f *fake) ReadRelationships(ctx context.Context, in *v1.ReadRelationshipsRequest, _ ...grpc.CallOption) (grpc.ServerStreamingClient[v1.ReadRelationshipsResponse], error) {
	if err := f.wait(ctx); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reads++
	flt := in.GetRelationshipFilter()
	var items []*v1.ReadRelationshipsResponse
	for _, t := range f.tuples {
		if !matches(flt, t) {
			continue
		}
		res := splitRef(t.Resource)
		sub := splitRef(t.Subject)
		items = append(items, &v1.ReadRelationshipsResponse{Relationship: &v1.Relationship{
			Resource: &v1.ObjectReference{ObjectType: res[0], ObjectId: res[1]},
			Relation: t.Relation,
			Subject:  &v1.SubjectReference{Object: &v1.ObjectReference{ObjectType: sub[0], ObjectId: sub[1]}},
		}})
	}
	return &stream[v1.ReadRelationshipsResponse]{items: items}, nil
}

func (f *fake) WriteRelationships(ctx context.Context, in *v1.WriteRelationshipsRequest, _ ...grpc.CallOption) (*v1.WriteRelationshipsResponse, error) {
	if err := f.wait(ctx); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writes = append(f.writes, in)
	for _, u := range in.GetUpdates() {
		r := u.GetRelationship()
		t := spicedb.Tuple{
			Resource: r.GetResource().GetObjectType() + ":" + r.GetResource().GetObjectId(),
			Relation: r.GetRelation(),
			Subject:  r.GetSubject().GetObject().GetObjectType() + ":" + r.GetSubject().GetObject().GetObjectId(),
		}
		i := slices.Index(f.tuples, t)
		switch u.GetOperation() {
		case v1.RelationshipUpdate_OPERATION_TOUCH:
			if i < 0 {
				f.tuples = append(f.tuples, t)
			}
		case v1.RelationshipUpdate_OPERATION_DELETE:
			if i >= 0 {
				f.tuples = slices.Delete(f.tuples, i, i+1)
			}
		case v1.RelationshipUpdate_OPERATION_UNSPECIFIED, v1.RelationshipUpdate_OPERATION_CREATE:
			// The Syncer writes neither.
		}
	}
	return &v1.WriteRelationshipsResponse{}, nil
}

// wait returns the error every call fails with, waiting for ctx first when
// the fake blocks.
func (f *fake) wait(ctx context.Context) error {
	if f.block {
		<-ctx.Done()
		return fmt.Errorf("fake SpiceDB: %w", ctx.Err())
	}
	return f.err
}

// matches reports whether t is a relationship flt selects.
func matches(flt *v1.RelationshipFilter, t spicedb.Tuple) bool {
	res, sub := splitRef(t.Resource), splitRef(t.Subject)
	switch {
	case flt.GetResourceType() != "" && res[0] != flt.GetResourceType():
		return false
	case flt.GetOptionalResourceId() != "" && res[1] != flt.GetOptionalResourceId():
		return false
	case flt.GetOptionalRelation() != "" && t.Relation != flt.GetOptionalRelation():
		return false
	case flt.GetOptionalSubjectFilter() != nil && sub[0] != flt.GetOptionalSubjectFilter().GetSubjectType():
		return false
	}
	return true
}

// splitRef splits "type:id".
func splitRef(s string) [2]string {
	for i := range len(s) {
		if s[i] == ':' {
			return [2]string{s[:i], s[i+1:]}
		}
	}
	return [2]string{s, ""}
}

// has reports whether the fake holds t.
func (f *fake) has(t spicedb.Tuple) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Contains(f.tuples, t)
}

// held returns the relationships the fake holds, in zed's notation, sorted.
func (f *fake) held() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, 0, len(f.tuples))
	for _, t := range f.tuples {
		out = append(out, t.String())
	}
	slices.Sort(out)
	return out
}
