package ui

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"

	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1/valuationv1connect"
)

// clientEventsRefresh is how long the UI keeps the core's answer about the client
// events before it asks again, so that a change the owner makes on one replica
// reaches the others within this time and a page costs the core no call of its own.
const clientEventsRefresh = 10 * time.Second

// clientEventsSwitch is whether the pages report client events (D-227), as the
// core stores it: the owner's setting, on until they turn it off. It is read
// without a session, since the login page reports too, and kept for
// clientEventsRefresh. When the core can't be asked it keeps the last answer, and
// with none it says off, so that a UI that can't tell collects nothing. The
// operator's config can still turn the reports off for the install, whatever the
// owner chose.
type clientEventsSwitch struct {
	settings valuationv1connect.SettingsServiceClient
	now      func() time.Time
	log      *slog.Logger
	// operatorOff is the plugin's clientEvents: false (Config.DisableClientEvents).
	operatorOff bool

	mu    sync.Mutex
	on    bool
	known bool
	at    time.Time
}

// enabled reports whether a page loads the reporting script and the endpoint
// keeps what it is sent.
func (c *clientEventsSwitch) enabled(ctx context.Context) bool {
	if c.operatorOff {
		return false
	}
	c.mu.Lock()
	if c.known && c.now().Sub(c.at) < clientEventsRefresh {
		on := c.on
		c.mu.Unlock()
		return on
	}
	c.mu.Unlock()
	return c.refresh(ctx)
}

// refresh asks the core and keeps its answer; on a failure it keeps the last
// answer for another interval, or off when there is none.
func (c *clientEventsSwitch) refresh(ctx context.Context) bool {
	resp, err := c.settings.GetClientEvents(ctx, connect.NewRequest(&valuationv1.GetClientEventsRequest{}))
	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		c.log.WarnContext(ctx, "client events setting could not be read", "code", connect.CodeOf(err).String())
		if ctx.Err() == nil {
			c.at = c.now() // a request that went away says nothing of the core
		}
		return c.known && c.on
	}
	c.on, c.known, c.at = resp.Msg.GetEnabled(), true, c.now()
	return c.on
}

// set keeps the owner's own change at once, on this replica.
func (c *clientEventsSwitch) set(on bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.on, c.known, c.at = on, true, c.now()
}

// clientEventsSetting is the Settings page's section on client events (D-227).
type clientEventsSetting struct {
	On bool
	// Forced is true when the install's config turns them off, which the owner's
	// choice does not override.
	Forced bool
}

// loadClientEvents reads the setting for the Settings page from the core, not from
// what is kept, so that the owner sees what is stored.
func (s *server) loadClientEvents(r *http.Request) *clientEventsSetting {
	if s.events.operatorOff {
		return &clientEventsSetting{Forced: true}
	}
	return &clientEventsSetting{On: s.events.refresh(r.Context())}
}

// setClientEvents turns the pages' client events on or off for the owner
// (SetClientEvents); the core enters the change in the audit log.
func (s *server) setClientEvents(w http.ResponseWriter, r *http.Request, sess session) {
	on := r.PostFormValue("enabled") == "1"
	_, err := s.settings.SetClientEvents(r.Context(), connect.NewRequest(&valuationv1.SetClientEventsRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Enabled: on,
	}))
	if err != nil {
		s.actionFailed(w, r, sess, "SetClientEvents", err)
		return
	}
	s.events.set(on)
	http.Redirect(w, r, "/settings#client-events-title", http.StatusSeeOther)
}
