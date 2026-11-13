// Package spicedb is the identity context's SpiceDB adapter (docs/DESIGN.md
// D-163, D-164): Authorizer answers identityapp.Authorizer from the
// relationships SpiceDB holds, over authz/schema.zed, and Syncer writes the
// relationships the core owns, the household's and each account's household
// and sleeve, from the portfolio's registry.
//
// Both fail closed: a SpiceDB error, a timeout or an ID it can't take is
// never an answer to allow. SpiceDB holds relationships only; the ledger and
// the portfolio stay the sources of which accounts exist and which sleeve
// each is in, and Syncer.Reconcile repairs SpiceDB from them.
package spicedb
