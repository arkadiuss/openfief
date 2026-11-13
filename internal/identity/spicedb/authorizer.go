package spicedb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"google.golang.org/grpc"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
)

// DefaultTimeout bounds one call to SpiceDB: a check that takes longer is
// refused.
const DefaultTimeout = 3 * time.Second

// ErrUnknownHolder reports a caller that is no holder, which may do nothing.
var ErrUnknownHolder = errors.New("no such holder")

// Permissions is what the adapter uses of SpiceDB's permissions service;
// authzed-go's PermissionsServiceClient implements it.
type Permissions interface {
	CheckPermission(ctx context.Context, in *v1.CheckPermissionRequest, opts ...grpc.CallOption) (*v1.CheckPermissionResponse, error)
	LookupResources(ctx context.Context, in *v1.LookupResourcesRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[v1.LookupResourcesResponse], error)
	ReadRelationships(ctx context.Context, in *v1.ReadRelationshipsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[v1.ReadRelationshipsResponse], error)
	WriteRelationships(ctx context.Context, in *v1.WriteRelationshipsRequest, opts ...grpc.CallOption) (*v1.WriteRelationshipsResponse, error)
}

// Schemas is the part of SpiceDB's schema service the Syncer needs; authzed-go's
// SchemaServiceClient implements it.
type Schemas interface {
	WriteSchema(ctx context.Context, in *v1.WriteSchemaRequest, opts ...grpc.CallOption) (*v1.WriteSchemaResponse, error)
}

// fullyConsistent has SpiceDB answer from its newest data, not a cached
// snapshot: a grant just written, or revoked, counts at once. The household
// is small, so the cost is too.
func fullyConsistent() *v1.Consistency {
	return &v1.Consistency{Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true}}
}

// Option configures an Authorizer or a Syncer.
type Option func(*options)

type options struct {
	timeout time.Duration
	// schemas and schema, when set, make a Syncer write the schema before
	// it reconciles.
	schemas Schemas
	schema  string
}

// WithTimeout bounds each call to SpiceDB by d instead of DefaultTimeout.
func WithTimeout(d time.Duration) Option {
	return func(o *options) { o.timeout = d }
}

// WithSchema makes a Syncer write schema through schemas at the start of every
// Reconcile, so SpiceDB always holds the model this build of the core was
// written against (D-164). Writing the schema it already holds changes
// nothing.
func WithSchema(schemas Schemas, schema string) Option {
	return func(o *options) { o.schemas, o.schema = schemas, schema }
}

func newOptions(opts []Option) options {
	o := options{timeout: DefaultTimeout}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// Authorizer answers identityapp.Authorizer from SpiceDB (D-164): whether a
// caller holds the account's view or change permission in authz/schema.zed.
// It fails closed: an error, a timeout, a caller or ID SpiceDB can't take or
// an answer that depends on a caveat (the schema has none) is a refusal,
// returned with its error, never the owner's view.
type Authorizer struct {
	perms   Permissions
	timeout time.Duration
}

var (
	_ identityapp.Authorizer    = (*Authorizer)(nil)
	_ identityapp.AccountLister = (*Authorizer)(nil)
)

// NewAuthorizer returns the Authorizer over perms.
func NewAuthorizer(perms Permissions, opts ...Option) *Authorizer {
	o := newOptions(opts)
	return &Authorizer{perms: perms, timeout: o.timeout}
}

// subject returns the SpiceDB subject of c: the owner is the one person
// there is, a script token, a view plugin and a demo session each their own
// ID. A caller that is no holder fails with ErrUnknownHolder.
func subject(c identity.Caller) (*v1.SubjectReference, error) {
	var typ, id string
	switch c.Holder {
	case identity.HolderOwner:
		typ, id = typePerson, ownerID
	case identity.HolderScriptToken:
		typ = typeScriptToken
	case identity.HolderViewPlugin:
		typ = typeViewPlugin
	case identity.HolderDemoSession:
		typ = typeDemoSession
	default:
		return nil, ErrUnknownHolder
	}
	if id == "" {
		var err error
		if id, err = ObjectID(c.ID); err != nil {
			return nil, fmt.Errorf("%s: %w", c.Holder, err)
		}
	}
	return &v1.SubjectReference{Object: &v1.ObjectReference{ObjectType: typ, ObjectId: id}}, nil
}

// MayAct reports whether c holds permission a on account in SpiceDB. A
// caller that is no holder, or an action that is none, may do nothing, with
// no call to SpiceDB. Any error is a refusal: the answer is false with the
// error.
func (z *Authorizer) MayAct(ctx context.Context, c identity.Caller, a identity.Action, account string) (bool, error) {
	if a != identity.ActionView && a != identity.ActionChange {
		return false, nil
	}
	subj, err := subject(c)
	if errors.Is(err, ErrUnknownHolder) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("authorizing %s: %w", a, err)
	}
	id, err := ObjectID(account)
	if err != nil {
		return false, fmt.Errorf("authorizing %s of an account: %w", a, err)
	}
	ctx, cancel := context.WithTimeout(ctx, z.timeout)
	defer cancel()
	resp, err := z.perms.CheckPermission(ctx, &v1.CheckPermissionRequest{
		Consistency: fullyConsistent(),
		Resource:    &v1.ObjectReference{ObjectType: typeAccount, ObjectId: id},
		Permission:  a.String(),
		Subject:     subj,
	})
	if err != nil {
		return false, fmt.Errorf("checking %s of an account in SpiceDB: %w", a, err)
	}
	return resp.GetPermissionship() == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, nil
}

// ActableAccounts returns the IDs of the accounts c holds permission a on,
// in one LookupResources call. Any error, or a caller or action that can't
// hold any, returns none with the error, so a list is never longer than
// what each account's MayAct would allow.
func (z *Authorizer) ActableAccounts(ctx context.Context, c identity.Caller, a identity.Action) (map[string]struct{}, error) {
	if a != identity.ActionView && a != identity.ActionChange {
		return map[string]struct{}{}, nil
	}
	subj, err := subject(c)
	if errors.Is(err, ErrUnknownHolder) {
		return map[string]struct{}{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("listing accounts to %s: %w", a, err)
	}
	ctx, cancel := context.WithTimeout(ctx, z.timeout)
	defer cancel()
	stream, err := z.perms.LookupResources(ctx, &v1.LookupResourcesRequest{
		Consistency:        fullyConsistent(),
		ResourceObjectType: typeAccount,
		Permission:         a.String(),
		Subject:            subj,
	})
	if err != nil {
		return nil, fmt.Errorf("looking up accounts to %s in SpiceDB: %w", a, err)
	}
	return collectAccounts(stream, a)
}

// collectAccounts reads the accounts a LookupResources stream holds, those
// whose permission is certain, decoding their IDs.
func collectAccounts(stream grpc.ServerStreamingClient[v1.LookupResourcesResponse], a identity.Action) (map[string]struct{}, error) {
	out := map[string]struct{}{}
	for {
		r, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return out, nil
		}
		if err != nil {
			return nil, fmt.Errorf("looking up accounts to %s in SpiceDB: %w", a, err)
		}
		if r.GetPermissionship() != v1.LookupPermissionship_LOOKUP_PERMISSIONSHIP_HAS_PERMISSION {
			continue
		}
		id, err := idOf(r.GetResourceObjectId())
		if err != nil {
			return nil, fmt.Errorf("looking up accounts to %s in SpiceDB: %w", a, err)
		}
		out[id] = struct{}{}
	}
}
