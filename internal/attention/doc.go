// Package attention is the attention context: the attention items that ask
// the owner to look at something, in one append-only store
// (docs/DESIGN.md D-130, D-131).
//
// Every check that needs the owner raises an attention item: its kind,
// such as a flagged balance assertion, and its subject, such as the
// assertion, name what it is about, and at most one item stands for each
// kind and subject. An item is open, acknowledged or resolved, and its
// history is a list of changes, each appended and never changed in place.
//
// A check reports the conditions it finds now (Condition); Observe and
// Reconcile turn what it found into changes: a new condition raises an
// item, a condition whose detail changed opens its item again, and a
// condition that cleared resolves its item. The owner acknowledges an open
// item (Acknowledge) and resolves one by hand only where no other page
// resolves it (Resolve, OwnerResolves).
//
// Everything here is a pure function over values.
package attention
