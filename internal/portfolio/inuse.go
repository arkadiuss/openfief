package portfolio

import "errors"

// Errors deleting a sleeve or kind something refers to (D-120): the owner
// deletes only a sleeve with no accounts and a kind no sleeve uses.
var (
	// ErrSleeveHasAccounts reports a sleeve that has accounts.
	ErrSleeveHasAccounts = errors.New("the sleeve has accounts")
	// ErrKindInUse reports a sleeve kind a sleeve uses.
	ErrKindInUse = errors.New("a sleeve uses the kind")
)
