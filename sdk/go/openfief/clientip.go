package openfief

// ClientIPHeader is the request header in which the core hands a view plugin the
// network address of the browser it proxies (D-81, D-208), such as 198.51.100.7
// or 2001:db8::1. The core sets it on every request it proxies, from the
// connection, or from the header of the reverse proxy in front of it when the
// core is configured to trust that proxy's address, and drops any value the
// browser sent in it; it leaves the header out when it can't tell the address.
// A view plugin may therefore believe it, since only the core can reach it
// (RequireCore), and uses it for what it is for: counting, never authorizing.
const ClientIPHeader = "X-Openfief-Client-Ip"
