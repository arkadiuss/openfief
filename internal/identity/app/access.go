package app

import (
	"context"

	"github.com/arkadiuss/openfief/internal/identity"
)

// Authorizer is the one interface every read path and owner service of the
// core asks whether a caller may act on an account (D-163). account is the
// account's ID. A false answer with a nil error refuses; an error means it
// can't tell now, and the caller is refused too. Until SpiceDB answers it
// (D-164), Policy is its only implementation.
type Authorizer interface {
	MayAct(ctx context.Context, c identity.Caller, a identity.Action, account string) (bool, error)
}

// AccountLister is what an Authorizer may offer besides MayAct: the accounts
// a caller may act on, all in one call, for a list or a total over every
// account (D-163). Its answer is exactly what MayAct would give for each
// account; on an error it returns none, and the caller is refused.
type AccountLister interface {
	// ActableAccounts returns the IDs of the accounts c may do a on.
	ActableAccounts(ctx context.Context, c identity.Caller, a identity.Action) (map[string]struct{}, error)
}

// Policy is the Authorizer until SpiceDB answers (D-163): the owner may act
// on every account, and every other holder of read access may view every
// account (identity.MayAct).
type Policy struct{}

var _ Authorizer = Policy{}

// MayAct implements Authorizer with identity.MayAct, for every account.
func (Policy) MayAct(_ context.Context, c identity.Caller, a identity.Action, _ string) (bool, error) {
	return identity.MayAct(c, a), nil
}

// callerKey is the context key of the caller a request was authenticated
// as.
type callerKey struct{}

// WithCaller returns ctx carrying c, the caller its request was
// authenticated as: the read API's interceptor puts it there before a
// handler runs.
func WithCaller(ctx context.Context, c identity.Caller) context.Context {
	return context.WithValue(ctx, callerKey{}, c)
}

// CallerFrom returns the caller WithCaller put in ctx, or the zero Caller,
// who may do nothing, when none is there.
func CallerFrom(ctx context.Context) identity.Caller {
	c, ok := ctx.Value(callerKey{}).(identity.Caller)
	if !ok {
		return identity.Caller{}
	}
	return c
}
