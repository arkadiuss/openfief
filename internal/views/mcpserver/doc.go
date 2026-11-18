// Package mcpserver serves MCP, the Model Context Protocol, at /mcp on the
// core's browser port (D-213): the Ask box's read-only answerers as tools for
// an external agent, which authenticates with a script token the owner allowed
// to use MCP. The token's grants limit what every tool sees, nothing it offers
// changes state, and each call is entered in the audit log as usage only. With
// the one switch for agents and models, or the MCP toggle under it, off, no
// endpoint answers (D-212).
//
// The package is the protocol's adapter, over the official Go SDK's
// Streamable HTTP handler in its stateless mode. It decides who may talk and
// records what they did; the tools, the grants and the audit log stay the
// ports' (ports.go), which cmd/core wires to the UI's toolbox, the read API
// and the identity context.
package mcpserver
