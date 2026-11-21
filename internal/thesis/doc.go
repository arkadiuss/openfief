// Package thesis is the thesis bounded context: draft theses (docs/DESIGN.md
// D-229, glossary: Thesis), the statement an agent on an offensive sleeve
// would act on, its instruments, its horizon, the conditions that would
// invalidate it and its sources.
//
// A draft is the owner's and no one else's: it is T1 data (D-206), it is not
// on the read API (D-137), and no grant reaches it (D-209). It has no agent,
// no fief and no lease, and no field for a lease's size: the control plane
// starts a fief from it later, as a proposal the owner approves (GR-6).
//
// The package is pure: no I/O, no clock, no logging. A draft's text is the
// owner's data and is never part of an error (GR-12).
package thesis
