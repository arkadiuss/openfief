package identity

// DemoSessionsAvailable reports whether the demo authenticator issues
// sessions (D-68): only when the core runs with demo sessions on
// (sessions) and the database carries the demo marker (marked).
func DemoSessionsAvailable(sessions, marked bool) bool {
	return sessions && marked
}

// DemoLoading reports whether a demo deployment still waits for its
// dataset: the core runs with demo sessions on (sessions), but the
// database doesn't carry the demo marker yet (marked). Until the demo
// loader writes it, the front door says the demo data is loading, and
// the core refuses what the first-run wizard would write, which would
// make the loader refuse the database as a real one.
func DemoLoading(sessions, marked bool) bool {
	return sessions && !marked
}
