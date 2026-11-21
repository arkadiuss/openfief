package identity

import "fmt"

// Action is what a caller asks to do with an account (D-163): view it, as
// every read path does, or change it, as the owner services do.
type Action int

// The actions on an account. The zero Action is none, and nobody may do it.
const (
	// ActionView reads an account and everything recorded in it.
	ActionView Action = iota + 1
	// ActionChange changes an account or what is recorded in it.
	ActionChange
)

// String returns the action's name: "view" or "change", the permissions
// of the account definition in authz/schema.zed (D-164).
func (a Action) String() string {
	switch a {
	case ActionView:
		return "view"
	case ActionChange:
		return "change"
	}
	return fmt.Sprintf("Action(%d)", int(a))
}

// Holder is who a caller is to the core: the owner, or one of the other
// holders of read access (D-137).
type Holder int

// The holders. The zero Holder is nobody.
const (
	// HolderOwner is the owner, through an owner-scoped session.
	HolderOwner Holder = iota + 1
	// HolderScriptToken is one of the owner's script tokens (D-27).
	HolderScriptToken
	// HolderViewPlugin is a view plugin, by its projected ServiceAccount
	// token (D-27).
	HolderViewPlugin
	// HolderDemoSession is a read-scoped session from the demo
	// authenticator.
	HolderDemoSession
	// HolderViewer is a viewer (D-65) in their read-scoped session; the
	// caller's ID is the viewer's person ID, whose grants say what they
	// view (D-209).
	HolderViewer
)

// String returns the holder's name: "owner", or the subject type
// authz/schema.zed gives the other holders, "script_token", "view_plugin"
// or "demo_session".
func (h Holder) String() string {
	switch h {
	case HolderOwner:
		return "owner"
	case HolderScriptToken:
		return "script_token"
	case HolderViewPlugin:
		return "view_plugin"
	case HolderDemoSession:
		return "demo_session"
	case HolderViewer:
		return "viewer"
	}
	return fmt.Sprintf("Holder(%d)", int(h))
}

// Caller is who asks the core to act on an account: a holder, and its ID
// among holders of its kind, such as a session's or a script token's ID or
// a view plugin's name. The zero Caller is nobody.
type Caller struct {
	Holder Holder
	ID     string
}

// MayAct reports whether c may do a on an account under the rule that holds
// until SpiceDB answers (D-163, D-164): the owner may view and change every
// account; every other holder of read access (D-137) may view every account
// and change none; nobody else may do anything. A viewer (D-65) holds only
// what their grants name, which this rule doesn't know, so it refuses them
// everything: their views are decided by their grants, in SpiceDB or, under
// the policy, in app.GrantPolicy.
func MayAct(c Caller, a Action) bool {
	switch c.Holder {
	case HolderOwner:
		return a == ActionView || a == ActionChange
	case HolderScriptToken, HolderViewPlugin, HolderDemoSession:
		return a == ActionView
	case HolderViewer:
		return false
	}
	return false
}

// MayReadQuotes reports whether c may read market quotes (D-226): public data
// fetched through the owner's market-data connector, tied to no account, so no
// grant decides it and only who holds the token does. The owner, a script
// token and a viewer may; a view plugin may not, since its manifest asks for
// no quotes (Ghostfolio's does not), and a demo session never may, since the
// demo answers from its synthetic data alone.
func MayReadQuotes(c Caller) bool {
	switch c.Holder {
	case HolderOwner, HolderScriptToken, HolderViewer:
		return true
	case HolderViewPlugin, HolderDemoSession:
		return false
	}
	return false
}
