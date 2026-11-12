package spicedb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

// errNoType reports an object reference without a type.
var errNoType = errors.New("no type")

// accountRelations is how many relationships an account has at most: its
// household and its sleeve.
const accountRelations = 2

// batchSize is how many relationship updates one WriteRelationships call
// carries, well under SpiceDB's default limit of 1,000.
const batchSize = 500

// Tuple is one relationship the Syncer owns: Resource#Relation@Subject, each
// side as "type:objectID" with the ID already an ObjectID.
type Tuple struct {
	Resource, Relation, Subject string
}

// String returns the tuple in zed's notation, for a log or a test failure.
func (t Tuple) String() string { return t.Resource + "#" + t.Relation + "@" + t.Subject }

// resourceID returns the ID of the tuple's resource.
func (t Tuple) resourceID() string {
	_, id, _ := strings.Cut(t.Resource, ":")
	return id
}

// Desired is what the core's own contexts say the relationships the Syncer
// owns must be: the accounts the ledger holds, the sleeve each account in
// one is in (the portfolio's), and whether demo sessions exist.
type Desired struct {
	// Accounts are the IDs of every account.
	Accounts []string
	// Sleeves maps an account ID to its sleeve's ID; an account in none is
	// absent.
	Sleeves map[string]string
	// DemoSessions is whether every demo session holds the household grant
	// (D-68, D-137).
	DemoSessions bool
}

// Syncer writes the relationships the core owns to SpiceDB: the household's
// owner and the demo sessions' household grant, and each account's
// household and sleeve (D-164). It never writes or removes a grant (a
// viewer relation of an account or sleeve, a household reader of a person,
// script token or view plugin): the owner's grants are theirs (D-209).
//
// SpiceDB is written after the database commits, in a call of its own, so a
// failure between them leaves the database ahead: AccountPlaced and
// AccountUnplaced return the error, and Reconcile, which the core runs at
// start, repairs any drift.
type Syncer struct {
	perms   Permissions
	timeout time.Duration
	// mu serializes the Syncer's read-then-write sequences, so a placement
	// can't land between Reconcile's read and its write and be undone.
	mu sync.Mutex
}

// NewSyncer returns the Syncer over perms.
func NewSyncer(perms Permissions, opts ...Option) *Syncer {
	o := newOptions(opts)
	return &Syncer{perms: perms, timeout: o.timeout}
}

// accountTuples returns the relationships account has: its household, and
// its sleeve when it is in one.
func accountTuples(account, sleeve string) ([]Tuple, error) {
	a, err := ObjectID(account)
	if err != nil {
		return nil, fmt.Errorf("account: %w", err)
	}
	res := typeAccount + ":" + a
	out := make([]Tuple, 0, accountRelations)
	out = append(out, Tuple{Resource: res, Relation: relHousehold, Subject: typeHousehold + ":" + householdID})
	if sleeve == "" {
		return out, nil
	}
	s, err := ObjectID(sleeve)
	if err != nil {
		return nil, fmt.Errorf("sleeve: %w", err)
	}
	return append(out, Tuple{Resource: res, Relation: relSleeve, Subject: typeSleeve + ":" + s}), nil
}

// householdTuples returns the household relationships the Syncer owns: its
// owner, and the demo sessions' grant when they exist.
func householdTuples(demo bool) []Tuple {
	res := typeHousehold + ":" + householdID
	out := []Tuple{{Resource: res, Relation: relOwner, Subject: typePerson + ":" + ownerID}}
	if demo {
		out = append(out, Tuple{Resource: res, Relation: relReader, Subject: typeDemoSession + ":" + wildcardID})
	}
	return out
}

// wanted returns every relationship d says the Syncer owns.
func wanted(d Desired) ([]Tuple, error) {
	out := householdTuples(d.DemoSessions)
	for _, a := range d.Accounts {
		ts, err := accountTuples(a, d.Sleeves[a])
		if err != nil {
			return nil, fmt.Errorf("account %q: %w", a, err)
		}
		out = append(out, ts...)
	}
	return out, nil
}

// Changes are what to write to SpiceDB: the relationships to touch and those
// to delete.
type Changes struct {
	Touch, Delete []Tuple
}

// Plan returns what to write so that have, the relationships the Syncer owns
// that SpiceDB holds, become want: every wanted one not held is touched,
// every held one not wanted deleted. Duplicates in either count once.
func Plan(have, want []Tuple) Changes {
	var c Changes
	h := make(map[Tuple]struct{}, len(have))
	for _, t := range have {
		h[t] = struct{}{}
	}
	w := make(map[Tuple]struct{}, len(want))
	for _, t := range want {
		w[t] = struct{}{}
		if _, ok := h[t]; !ok {
			c.Touch = append(c.Touch, t)
			h[t] = struct{}{} // once
		}
	}
	for _, t := range have {
		if _, ok := w[t]; !ok {
			c.Delete = append(c.Delete, t)
			w[t] = struct{}{} // once
		}
	}
	return c
}

// AccountPlaced writes account's household and puts it in sleeve, in one
// atomic write that also removes the sleeve it was in; an empty sleeve only
// takes it out of one, as AccountUnplaced does. It is idempotent.
func (s *Syncer) AccountPlaced(ctx context.Context, account, sleeve string) error {
	want, err := accountTuples(account, sleeve)
	if err != nil {
		return fmt.Errorf("placing an account in SpiceDB: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	have, err := s.read(ctx, &v1.RelationshipFilter{ResourceType: typeAccount, OptionalResourceId: want[0].resourceID(), OptionalRelation: relSleeve})
	if err != nil {
		return fmt.Errorf("placing an account in SpiceDB: %w", err)
	}
	// The account's household is only touched, never removed, by a move.
	var stale []Tuple
	for _, t := range have {
		if !slices.Contains(want, t) {
			stale = append(stale, t)
		}
	}
	if err := s.write(ctx, want, stale); err != nil {
		return fmt.Errorf("placing an account in SpiceDB: %w", err)
	}
	return nil
}

// AccountUnplaced takes account out of its sleeve, which is all that
// removing an account from its sleeve does: the account keeps its household
// until Reconcile finds it deleted. Its household stays so that the owner
// never loses an account the ledger still holds. It is idempotent.
func (s *Syncer) AccountUnplaced(ctx context.Context, account string) error {
	id, err := ObjectID(account)
	if err != nil {
		return fmt.Errorf("unplacing an account in SpiceDB: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	have, err := s.read(ctx, &v1.RelationshipFilter{ResourceType: typeAccount, OptionalResourceId: id, OptionalRelation: relSleeve})
	if err != nil {
		return fmt.Errorf("unplacing an account in SpiceDB: %w", err)
	}
	if err := s.write(ctx, nil, have); err != nil {
		return fmt.Errorf("unplacing an account in SpiceDB: %w", err)
	}
	return nil
}

// Reconcile makes the relationships the Syncer owns what load says: it writes
// the missing ones and removes the ones load no longer wants, such as those
// of a deleted account, leaving every grant alone. load runs under the
// Syncer's lock, so a placement committed after it ran is written after
// Reconcile, not undone by it. Reconcile is idempotent, and run at start it
// repairs what a failed write after a commit left.
func (s *Syncer) Reconcile(ctx context.Context, load func(context.Context) (Desired, error)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, err := load(ctx)
	if err != nil {
		return fmt.Errorf("reconciling SpiceDB: loading what the core holds: %w", err)
	}
	want, err := wanted(d)
	if err != nil {
		return fmt.Errorf("reconciling SpiceDB: %w", err)
	}
	var have []Tuple
	for _, f := range ownedFilters() {
		ts, err := s.read(ctx, f)
		if err != nil {
			return fmt.Errorf("reconciling SpiceDB: %w", err)
		}
		have = append(have, ts...)
	}
	c := Plan(have, want)
	if err := s.write(ctx, c.Touch, c.Delete); err != nil {
		return fmt.Errorf("reconciling SpiceDB: %w", err)
	}
	return nil
}

// ownedFilters are the relationships Reconcile reads: every one of a kind
// the Syncer owns, and nothing of a grant.
func ownedFilters() []*v1.RelationshipFilter {
	return []*v1.RelationshipFilter{
		{ResourceType: typeAccount, OptionalRelation: relHousehold},
		{ResourceType: typeAccount, OptionalRelation: relSleeve},
		{ResourceType: typeHousehold, OptionalRelation: relOwner},
		{
			ResourceType: typeHousehold, OptionalRelation: relReader,
			OptionalSubjectFilter: &v1.SubjectFilter{SubjectType: typeDemoSession},
		},
	}
}

// read returns the relationships matching f.
func (s *Syncer) read(ctx context.Context, f *v1.RelationshipFilter) ([]Tuple, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	stream, err := s.perms.ReadRelationships(ctx, &v1.ReadRelationshipsRequest{Consistency: fullyConsistent(), RelationshipFilter: f})
	if err != nil {
		return nil, fmt.Errorf("reading relationships: %w", err)
	}
	var out []Tuple
	for {
		r, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return out, nil
		}
		if err != nil {
			return nil, fmt.Errorf("reading relationships: %w", err)
		}
		rel := r.GetRelationship()
		out = append(out, Tuple{
			Resource: ref(rel.GetResource().GetObjectType(), rel.GetResource().GetObjectId()),
			Relation: rel.GetRelation(),
			Subject:  ref(rel.GetSubject().GetObject().GetObjectType(), rel.GetSubject().GetObject().GetObjectId()),
		})
	}
}

// ref joins an object's type and ID.
func ref(typ, id string) string { return typ + ":" + id }

// write touches touch and deletes del, atomically when they fit in one call
// and in order, touches first, in batches of batchSize when they don't.
func (s *Syncer) write(ctx context.Context, touch, del []Tuple) error {
	updates := make([]*v1.RelationshipUpdate, 0, len(touch)+len(del))
	for _, t := range touch {
		u, err := update(v1.RelationshipUpdate_OPERATION_TOUCH, t)
		if err != nil {
			return err
		}
		updates = append(updates, u)
	}
	for _, t := range del {
		u, err := update(v1.RelationshipUpdate_OPERATION_DELETE, t)
		if err != nil {
			return err
		}
		updates = append(updates, u)
	}
	for batch := range slices.Chunk(updates, batchSize) {
		if err := s.writeUpdates(ctx, batch); err != nil {
			return err
		}
	}
	return nil
}

// writeUpdates writes updates in one call.
func (s *Syncer) writeUpdates(ctx context.Context, updates []*v1.RelationshipUpdate) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	if _, err := s.perms.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{Updates: updates}); err != nil {
		return fmt.Errorf("writing relationships: %w", err)
	}
	return nil
}

// update returns the RelationshipUpdate doing op to t.
func update(op v1.RelationshipUpdate_Operation, t Tuple) (*v1.RelationshipUpdate, error) {
	res, err := objectOf(t.Resource)
	if err != nil {
		return nil, err
	}
	subj, err := objectOf(t.Subject)
	if err != nil {
		return nil, err
	}
	return &v1.RelationshipUpdate{
		Operation: op,
		Relationship: &v1.Relationship{
			Resource: res,
			Relation: t.Relation,
			Subject:  &v1.SubjectReference{Object: subj},
		},
	}, nil
}

// objectOf parses "type:id".
func objectOf(s string) (*v1.ObjectReference, error) {
	typ, id, ok := strings.Cut(s, ":")
	if !ok {
		return nil, fmt.Errorf("object %q: %w", s, errNoType)
	}
	return &v1.ObjectReference{ObjectType: typ, ObjectId: id}, nil
}
