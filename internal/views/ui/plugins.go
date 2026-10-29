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

// trustLabel names how far a plugin's image is trusted (D-45, D-56).
func trustLabel(t viewsv1.SignatureTrust) string {
	switch t {
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_VERIFIED:
		return "Verified plugin"
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_OPENFIEF_CI:
		return "Signed by OpenFief's CI"
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_PUBLISHER:
		return "Signed by a catalog publisher"
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_OWNER:
		return "Signed by your own signer"
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_TRUSTED_SIGNER:
		return "Signed by a trusted signer"
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_REFUSED:
		return "Refused"
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_NOT_CHECKED:
		return "Not checked yet"
	case viewsv1.SignatureTrust_SIGNATURE_TRUST_UNSPECIFIED:
	}
	return "Unknown"
}
