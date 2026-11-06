// Package archtest holds OpenFief's architecture tests
// (docs/dev/code-organization.md#architecture-tests): rules over the whole
// repository's packages, imports, types, SQL, contracts and tests, which a
// linter that sees one file at a time cannot check. It is a development tool,
// never shipped.
//
// Load reads a repository with go/packages (types from the compiler's export
// data, so only the repository's own packages are type-checked from source)
// and walks the directories it names for tests, SQL and contracts. Each rule
// is a function of the loaded repository that returns its violations, each
// with a stable key: a package, file or function, never a line number.
// Compare matches them against the checked-in baseline (breaks of today,
// which may only shrink) and the exceptions (the few the rules allow, each
// with its reason); the test in this package runs every rule on the
// repository and fails on a new violation or a stale line.
package archtest
