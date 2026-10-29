package ui

import (
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// pluginList is the installed plugins as the read API lists them (D-46,
// D-56), or that the core can't read them now.
type pluginList struct {
	Plugins     []*viewsv1.Plugin
	Unavailable bool
}

// loadPlugins reads the installed plugins with the session's token. A core
// that can't read its cluster answers unavailable, which the page shows as
// a sentence; any other failure is returned.
func (s *server) loadPlugins(r *http.Request, sess session) (pluginList, error) {
	resp, err := s.read.ListPlugins(r.Context(), bearer(&viewsv1.ListPluginsRequest{}, sess))
	if connect.CodeOf(err) == connect.CodeUnavailable {
		return pluginList{Unavailable: true}, nil
	}
	if err != nil {
		return pluginList{}, fmt.Errorf("listing plugins: %w", err)
	}
	return pluginList{Plugins: resp.Msg.GetPlugins()}, nil
}

// pluginsPage shows the installed plugins to a session of either scope. It
// only shows them: plugins are installed and approved with the openfief CLI
// and the owner's kubeconfig, never through the UI (D-46).
func (s *server) pluginsPage(w http.ResponseWriter, r *http.Request, sess session) {
	list, err := s.loadPlugins(r, sess)
	if err != nil {
		s.coreFailed(w, r, "ListPlugins", err)
		return
	}
	s.show(w, r, http.StatusOK, "plugins", page{Title: "Plugins", Session: &sess, Data: list})
}

// trustBadge is one of the four badges the UI groups a plugin's signature
// trust into (D-87): its word and the tone of its pill.
type trustBadge struct {
	Label string
	// Tone is the status pill's tone: ok, neutral, warn or danger.
	Tone string
}

// badgeOf groups a plugin's signature trust into its badge (D-87):
// Verified when the catalog pins the reviewed version (D-44), Yours when
// one of the owner's signers signed it (D-45), Unreviewed for any other
// trusted signer, and Blocked when the operator refused it or hasn't
// checked it yet. A trust the UI doesn't know is Blocked too.
func badgeOf(t viewsv1.SignatureTrust) trustBadge {
	switch t {
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_VERIFIED:
		return trustBadge{Label: "Verified", Tone: "ok"}
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_OWNER:
		return trustBadge{Label: "Yours", Tone: "neutral"}
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_OPENFIEF_CI,
		viewsv1.SignatureTrust_SIGNATURE_TRUST_PUBLISHER,
		viewsv1.SignatureTrust_SIGNATURE_TRUST_TRUSTED_SIGNER:
		return trustBadge{Label: "Unreviewed", Tone: "warn"}
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_REFUSED,
		viewsv1.SignatureTrust_SIGNATURE_TRUST_NOT_CHECKED,
		viewsv1.SignatureTrust_SIGNATURE_TRUST_UNSPECIFIED:
	}
	return trustBadge{Label: "Blocked", Tone: "danger"}
}

// trustLabel names a plugin's signature trust in full, the detail under
// its badge (D-87): who signed the image, or why it is blocked.
func trustLabel(t viewsv1.SignatureTrust) string {
	switch t {
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_VERIFIED:
		return "In the catalog and reviewed"
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_OPENFIEF_CI:
		return "Signed by OpenFief's CI"
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_PUBLISHER:
		return "Signed by a catalog publisher"
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_OWNER:
		return "Signed by your own signer"
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_TRUSTED_SIGNER:
		return "Signed by a trusted signer"
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_REFUSED:
		return "Refused by the operator"
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_NOT_CHECKED:
		return "Not checked yet"
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_UNSPECIFIED:
	}
	return "Unknown"
}
