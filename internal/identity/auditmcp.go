package identity

import (
	"fmt"
	"slices"
	"strings"
)

// RefMCPTool is the MCP tool a script token called, by its name (D-213,
// D-211): a lower-case word of letters, digits and underscores, such as
// "allocation". The core writes the name of a tool it offers, never text a
// client sent.
const RefMCPTool RefKind = "mcp_tool"

// maxMCPToolLen is the longest MCP tool name an entry names.
const maxMCPToolLen = 48

// validMCPTool reports whether id has the shape of an MCP tool's name: one to
// maxMCPToolLen characters, a lower-case letter first, then lower-case
// letters, digits and underscores. An amount, a date or a secret does not
// have that shape.
func validMCPTool(id string) bool {
	if id == "" || len(id) > maxMCPToolLen || id[0] < 'a' || id[0] > 'z' || looksLikeSecret(id) {
		return false
	}
	return !strings.ContainsFunc(id, func(c rune) bool {
		return (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '_'
	})
}

// mcpOutcomes are the outcomes of an MCP call's entry.
func mcpOutcomes() []AuditOutcome {
	return []AuditOutcome{OutcomeOK, OutcomeRefused, OutcomeFailed}
}

// validateMCPEntry checks an MCP call's entry (D-213, D-211): a script token
// is the actor, the entry names the tool once and nothing else, and the
// outcome is ok, refused or failed. The one entry that names no tool is the
// refusal of a valid token the owner did not allow to use MCP: refused, with
// no reference.
func validateMCPEntry(e *AuditEntry) error {
	notAllowed := len(e.Refs) == 0 && e.Outcome == OutcomeRefused
	called := len(e.Refs) == 1 && e.Refs[0].Kind == RefMCPTool && slices.Contains(mcpOutcomes(), e.Outcome)
	if e.Actor.Kind != ActorScriptToken || (!notAllowed && !called) {
		return fmt.Errorf("%w: an %s entry is a script token's, names its tool and ends ok, refused or failed, or is a refusal naming none", ErrInvalidAuditEntry, e.Action)
	}
	return nil
}
