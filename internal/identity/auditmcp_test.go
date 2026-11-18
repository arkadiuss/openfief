package identity_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/identity"
)

// TestMCPCallEntry checks that an MCP call's entry (D-213, D-211) is a script
// token's, names its tool once and ends ok, refused or failed.
func TestMCPCallEntry(t *testing.T) {
	t.Parallel()
	token := identity.Actor{Kind: identity.ActorScriptToken, ID: auditTarget}
	tool := identity.Ref{Kind: identity.RefMCPTool, ID: "allocation"}
	tests := []struct {
		name    string
		actor   identity.Actor
		outcome identity.AuditOutcome
		refs    []identity.Ref
		wantErr bool
	}{
		{name: "a call that ran", actor: token, outcome: identity.OutcomeOK, refs: []identity.Ref{tool}},
		{name: "a call whose arguments were refused", actor: token, outcome: identity.OutcomeRefused, refs: []identity.Ref{tool}},
		{name: "a call whose data could not be read", actor: token, outcome: identity.OutcomeFailed, refs: []identity.Ref{tool}},
		{name: "a tool name with digits and underscores", actor: token, outcome: identity.OutcomeOK, refs: []identity.Ref{{Kind: identity.RefMCPTool, ID: "what_needs_me2"}}},
		{name: "the sent outcome of a model call", actor: token, outcome: identity.OutcomeSent, refs: []identity.Ref{tool}, wantErr: true},
		{name: "a token not allowed to use MCP is refused, naming no tool", actor: token, outcome: identity.OutcomeRefused},
		{name: "no tool and not a refusal", actor: token, outcome: identity.OutcomeOK, wantErr: true},
		{name: "no tool and failed", actor: token, outcome: identity.OutcomeFailed, wantErr: true},
		{name: "two tools", actor: token, outcome: identity.OutcomeOK, refs: []identity.Ref{tool, {Kind: identity.RefMCPTool, ID: "value"}}, wantErr: true},
		{name: "the token also named", actor: token, outcome: identity.OutcomeOK, refs: []identity.Ref{tool, {Kind: identity.RefScriptToken, ID: auditTarget}}, wantErr: true},
		{name: "the owner as actor", actor: identity.Actor{Kind: identity.ActorOwner, ID: auditSession}, outcome: identity.OutcomeOK, refs: []identity.Ref{tool}, wantErr: true},
		{name: "a token named by its secret", actor: identity.Actor{Kind: identity.ActorScriptToken, ID: identity.SecretPrefix + auditTarget}, outcome: identity.OutcomeOK, refs: []identity.Ref{tool}, wantErr: true},
		{name: "a token without an ID", actor: identity.Actor{Kind: identity.ActorScriptToken}, outcome: identity.OutcomeOK, refs: []identity.Ref{tool}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := identity.NewAuditEntry(auditAt(), tc.actor, identity.ActionMCPCall, tc.outcome, tc.refs...)
			if tc.wantErr {
				if !errors.Is(err, identity.ErrInvalidAuditEntry) {
					t.Fatalf("NewAuditEntry() error = %v, want %v", err, identity.ErrInvalidAuditEntry)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewAuditEntry() error = %v", err)
			}
		})
	}
}

// TestMCPToolRefIsAName checks that nothing but a tool's name is a reference
// to an MCP tool: the secrets and texts GR-12 keeps out of the log, an amount
// and a long text are refused.
func TestMCPToolRefIsAName(t *testing.T) {
	t.Parallel()
	for name, v := range rejectedValues() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			checkRefRejected(t, identity.RefMCPTool, name, v)
		})
	}
	for _, v := range []string{"Allocation", "_value", "1value", "what-needs-me", "show ", strings.Repeat("a", 49), "ofs_abc"} {
		if err := (identity.Ref{Kind: identity.RefMCPTool, ID: v}).Validate(); !errors.Is(err, identity.ErrInvalidAuditEntry) {
			t.Errorf("Ref{mcp_tool, %q}.Validate() = %v, want %v", v, err, identity.ErrInvalidAuditEntry)
		}
	}
	if k, err := identity.ParseRefKind(string(identity.RefMCPTool)); err != nil || k != identity.RefMCPTool {
		t.Errorf("ParseRefKind(mcp_tool) = %q, %v", k, err)
	}
}
