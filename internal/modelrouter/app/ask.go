package app

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/arkadiuss/openfief/internal/modelrouter"
)

// Askers checks the session of someone asking the Ask box a question (D-203):
// any valid session on an allowed origin, a read-scoped one included, since
// the demo's visitors ask too and what the model sees is limited by what the
// session may read. It returns who asked, as the audit entries name the actor:
// the owner or a viewer, by the session's ID; a viewer's question is the
// viewer's, never the owner's. cmd/core answers it over the identity context,
// with errors wrapping ErrNoSession.
type Askers interface {
	Asker(ctx context.Context, origin, sessionToken string) (Actor, error)
}

// Ask is the Ask box's door to the model router (D-203): for the Ask box's
// role alone, for the person whose session is valid. The tool loop runs in
// the built-in UI, which holds the read-only answerers the model may call
// and reads with the asking person's session; the router stays here, so
// every round goes through the switch, the route, the data tiers and the
// audit log.
//
// It holds no question or answer beyond a call's duration, and records and
// returns no text in an error (GR-12).
type Ask struct {
	askers   Askers
	settings SettingsStore
	router   *Router
	// questions and limits count each question to a model per session and per
	// address (D-208); nil questions or no limit lets every question through.
	questions QuestionStore
	limits    modelrouter.QuestionLimits
	// demo is whether the core is a public demo (D-68), which never gives
	// opinions (D-225, R-7).
	demo bool
}

// NewAsk returns the Ask box's door over router, with sessions checked by
// askers and the switch and route read from settings.
func NewAsk(askers Askers, settings SettingsStore, router *Router) *Ask {
	return &Ask{askers: askers, settings: settings, router: router}
}

// WithQuestionLimits returns a with each question to a model counted in store
// and held to limits per session and per address (D-208). Limits that set
// none leave the door as it was.
func (a *Ask) WithQuestionLimits(store QuestionStore, limits modelrouter.QuestionLimits) *Ask {
	out := *a
	out.questions, out.limits = store, limits
	return &out
}

// WithDemo returns a that knows whether the core is a public demo, where the
// Ask box never gives an opinion or suggests a thesis (D-225, R-7).
func (a *Ask) WithDemo(demo bool) *Ask {
	out := *a
	out.demo = demo
	return &out
}

// AdmitDecision is the Ask box's door's answer to a question that is about to
// go to a model.
type AdmitDecision string

// The decisions on a question.
const (
	// AdmitAllowed: the question may go to the model; it is counted.
	AdmitAllowed AdmitDecision = "allowed"
	// AdmitBudgetSpent: the day's budget is used up; nothing was counted.
	AdmitBudgetSpent AdmitDecision = "budget-spent"
	// AdmitSessionLimit: the session asked as many questions as it may.
	AdmitSessionLimit AdmitDecision = "session-limit"
	// AdmitAddressLimit: the address asked as many as it may.
	AdmitAddressLimit AdmitDecision = "address-limit"
)

// Admission is a decision with, when the question is not allowed, when the
// reason ends: the budget's next day, or the moment enough of the subject's
// earlier questions have left the window.
type Admission struct {
	Decision AdmitDecision
	ResetsAt time.Time
}

// Admit decides whether a question the Ask box is about to send a model may
// go, once per question and not per call of its tool loop, and counts it when
// it may. It does so before the first call: first the day's budget, which
// costs the person nothing when it is spent, then the limits per session and,
// when address is known, per address. address is the IP address of the asker
// as the core's front door found it, which this does not check beyond its
// form: an unparseable one counts as unknown and leaves the address limit
// out. An IPv6 address is counted as its /64, since a person with one has
// the whole network. It writes no audit entry: nothing is called, and a
// refused call is audited when one is made.
func (a *Ask) Admit(ctx context.Context, origin, sessionToken, address string) (Admission, error) {
	who, err := a.asker(ctx, origin, sessionToken)
	if err != nil {
		return Admission{}, err
	}
	state, err := a.router.BudgetState(ctx)
	if err != nil {
		return Admission{}, fmt.Errorf("reading the day's budget: %w", err)
	}
	if state.Spent {
		return Admission{Decision: AdmitBudgetSpent, ResetsAt: state.ResetsAt}, nil
	}
	if a.questions == nil || !a.limits.Enabled() {
		return Admission{Decision: AdmitAllowed}, nil
	}
	q := QuestionAsked{At: a.router.now(), Session: who.ID, Address: addressKey(address), Limits: a.limits}
	ruling, err := a.questions.Admit(ctx, &q)
	if err != nil {
		return Admission{}, fmt.Errorf("counting a question: %w", err)
	}
	switch {
	case ruling.Allowed:
		return Admission{Decision: AdmitAllowed}, nil
	case ruling.Subject == SubjectAddress:
		return Admission{Decision: AdmitAddressLimit, ResetsAt: ruling.ResetsAt}, nil
	}
	return Admission{Decision: AdmitSessionLimit, ResetsAt: ruling.ResetsAt}, nil
}

// addressKey returns the address a question is counted under: an IPv4 address
// as written in its canonical form, an IPv6 address as its /64 network, and ""
// for text that is no address.
func addressKey(s string) string {
	addr, err := netip.ParseAddr(strings.TrimSpace(s))
	if err != nil {
		return ""
	}
	addr = addr.Unmap().WithZone("")
	if addr.Is6() {
		prefix, err := addr.Prefix(ipv6Network)
		if err != nil {
			return ""
		}
		return prefix.String()
	}
	return addr.String()
}

// ipv6Network is the prefix length an IPv6 address is counted under.
const ipv6Network = 64

// Available reports whether the Ask box has a model for the person with the
// session: the one switch and the Ask box's toggle are on and its route has a
// model (D-212). The data tiers decide per request. It writes no audit entry.
func (a *Ask) Available(ctx context.Context, origin, sessionToken string) (bool, error) {
	if _, err := a.asker(ctx, origin, sessionToken); err != nil {
		return false, err
	}
	stored, err := a.settings.Settings(ctx)
	if err != nil {
		return false, fmt.Errorf("loading the models settings: %w", err)
	}
	return stored.HasModel(modelrouter.RoleAsk), nil
}

// Opinions says whether the Ask box may give opinions and suggest a thesis to
// the person with the session (D-225, D-229): only the owner's own session, on
// an install that is no demo, with the owner's setting on under the one switch
// and the Ask box's toggle (D-212), and a route with a model allowed for the
// question and cards (T0), a thesis's text (T1) and prices (T2) together.
// When the setting is on and only the route keeps opinions off, the status
// says so, and the box tells the owner why. It calls no model and writes no
// audit entry.
func (a *Ask) Opinions(ctx context.Context, origin, sessionToken string) (modelrouter.OpinionsStatus, error) {
	who, err := a.asker(ctx, origin, sessionToken)
	if err != nil {
		return modelrouter.OpinionsOff, err
	}
	stored, err := a.settings.Settings(ctx)
	if err != nil {
		return modelrouter.OpinionsOff, fmt.Errorf("loading the models settings: %w", err)
	}
	status, err := modelrouter.Opinions(&modelrouter.OpinionsInput{
		Settings: stored, Owner: who.Kind == ActorOwner, Demo: a.demo,
		RouteEligible: func() (bool, error) {
			permitErr := a.router.Permits(ctx, modelrouter.RoleAsk, modelrouter.OpinionsTiers(), modelrouter.Needs{Tools: true})
			if refusal(permitErr) {
				return false, nil
			}
			return permitErr == nil, permitErr
		},
	})
	if err != nil {
		return modelrouter.OpinionsOff, fmt.Errorf("deciding whether the Ask box gives opinions: %w", err)
	}
	return status, nil
}

// Permits reports whether a model of the Ask box's route may receive data of
// every one of tiers together, with tools (D-206), without calling any model
// or writing an audit entry. A tier that is none of T0, TF, T1 and T2 is an
// error wrapping ErrInvalidInput. A route that may not is false and no error:
// that is the answer, a refusal of the data and not a failure.
func (a *Ask) Permits(ctx context.Context, origin, sessionToken string, tiers []string) (bool, error) {
	if _, err := a.asker(ctx, origin, sessionToken); err != nil {
		return false, err
	}
	parsed := make([]modelrouter.DataTier, 0, len(tiers))
	for _, t := range tiers {
		tier, err := modelrouter.ParseDataTier(t)
		if err != nil {
			return false, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		parsed = append(parsed, tier)
	}
	set, err := modelrouter.NewDataTierSet(parsed)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	err = a.router.Permits(ctx, modelrouter.RoleAsk, set, modelrouter.Needs{Tools: true})
	if refusal(err) {
		return false, nil
	}
	return err == nil, err
}

// Generate asks the Ask box's models for the next message of req on behalf of
// the person with the session, through the router (Router.Generate): the
// router's refusals, audit failures and the exhaustion of the route come back
// as its own errors.
func (a *Ask) Generate(ctx context.Context, origin, sessionToken string, req *modelrouter.Request) (modelrouter.Result, error) {
	who, err := a.asker(ctx, origin, sessionToken)
	if err != nil {
		return modelrouter.Result{}, err
	}
	res, err := a.router.Generate(ctx, modelrouter.RoleAsk, who, req)
	if err != nil {
		return modelrouter.Result{}, fmt.Errorf("asking the Ask box's model: %w", err)
	}
	return res, nil
}

// asker returns who asked with the session of sessionToken on origin.
func (a *Ask) asker(ctx context.Context, origin, sessionToken string) (Actor, error) {
	who, err := a.askers.Asker(ctx, origin, sessionToken)
	if err != nil {
		return Actor{}, fmt.Errorf("the Ask box: %w", err)
	}
	return who, nil
}

// refusal reports whether err is the router saying the data, not the
// settings or the providers, may not go: no model of the route may receive it
// or serves what it asks.
func refusal(err error) bool {
	return errors.Is(err, modelrouter.ErrDataTierDenied) || errors.Is(err, modelrouter.ErrNoCapableModel)
}
