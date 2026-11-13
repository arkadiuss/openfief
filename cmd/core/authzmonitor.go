package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// authzProbeEvery is how often the monitor asks a SpiceDB that failed whether
// it answers again, so the way back is seen even with no request to see it.
const authzProbeEvery = 5 * time.Second

// The states of the authorizer's availability.
const (
	authzUnknown int32 = iota
	authzAvailable
	authzUnavailable
)

// authzMonitor watches whether the authorizer can answer, after the core
// started (D-164, D-211's spirit: closed, and said once). Readiness is the
// start's alone and never flips on an outage; instead every check that
// can't be answered fails closed with kernel.ErrAuthzUnavailable, which the
// APIs answer as 503, and the monitor logs each change of state once and
// exposes it as a metric. With the metrics pipeline not built (D-172), the
// metric is an instrument of OpenTelemetry's global meter, which records
// nothing until the core installs a provider.
type authzMonitor struct {
	log   *slog.Logger
	state atomic.Int32
	// probing is set while a goroutine waits for SpiceDB to answer again.
	probing atomic.Bool
	// ping asks SpiceDB whether it answers; nil without a probe.
	ping func(context.Context) error
	// probeCtx ends the probes with the core.
	probeCtx context.Context //nolint:containedctx // the monitor outlives every request and stops with the core
	every    time.Duration

	changes metric.Int64Counter
}

// newAuthzMonitor returns a monitor that logs to log and records to meter,
// probing with ping each time every has passed after a failure, until ctx
// ends.
func newAuthzMonitor(ctx context.Context, log *slog.Logger, meter metric.Meter, ping func(context.Context) error, every time.Duration) (*authzMonitor, error) {
	m := &authzMonitor{log: log, ping: ping, probeCtx: ctx, every: every}
	var err error
	m.changes, err = meter.Int64Counter("openfief.authz.availability.changes",
		metric.WithDescription("How often the authorization model became available or unavailable, by the state it changed to"))
	if err != nil {
		return nil, fmt.Errorf("authorization metric: %w", err)
	}
	up, err := meter.Int64ObservableGauge("openfief.authz.available",
		metric.WithDescription("1 while the authorization model answers, 0 while it can't; absent until the first check"))
	if err != nil {
		return nil, fmt.Errorf("authorization metric: %w", err)
	}
	_, err = meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		switch m.state.Load() {
		case authzAvailable:
			o.ObserveInt64(up, 1)
		case authzUnavailable:
			o.ObserveInt64(up, 0)
		}
		return nil
	}, up)
	if err != nil {
		return nil, fmt.Errorf("authorization metric: %w", err)
	}
	return m, nil
}

// observe records the outcome of one check or lookup: only an error that
// wraps kernel.ErrAuthzUnavailable is an outage, any other, a refusal or a
// bad request, is a working model.
func (m *authzMonitor) observe(err error) {
	if err != nil && errors.Is(err, kernel.ErrAuthzUnavailable) {
		if m.state.Swap(authzUnavailable) != authzUnavailable {
			m.log.Warn("authorization is unavailable: SpiceDB does not answer; every check is refused and the requests that need one get 503 until it does")
			m.changes.Add(m.probeCtx, 1, metric.WithAttributes(attribute.String("to", "unavailable")))
			m.startProbe()
		}
		return
	}
	if m.state.Swap(authzAvailable) == authzUnavailable {
		m.log.Info("authorization is available again: SpiceDB answers")
		m.changes.Add(m.probeCtx, 1, metric.WithAttributes(attribute.String("to", "available")))
	}
}

// startProbe starts, unless one runs, the goroutine that asks SpiceDB every
// m.every whether it answers, and reports the first answer.
func (m *authzMonitor) startProbe() {
	if m.ping == nil || !m.probing.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer m.probing.Store(false)
		t := time.NewTicker(m.every)
		defer t.Stop()
		for m.state.Load() == authzUnavailable {
			select {
			case <-m.probeCtx.Done():
				return
			case <-t.C:
			}
			if err := m.ping(m.probeCtx); err == nil {
				m.observe(nil)
			}
		}
	}()
}

// monitored is identityapp.Authorizer that tells its monitor each outcome.
type monitored struct {
	inner identityapp.Authorizer
	m     *authzMonitor
}

func (a monitored) MayAct(ctx context.Context, c identity.Caller, act identity.Action, account string) (bool, error) {
	ok, err := a.inner.MayAct(ctx, c, act, account)
	a.m.observe(err)
	return ok, err //nolint:wrapcheck // the wrapped authorizer's error, unchanged, is what callers match
}

// monitoredLister is monitored over an authorizer that lists accounts too.
type monitoredLister struct {
	monitored

	lister identityapp.AccountLister
}

func (a monitoredLister) ActableAccounts(ctx context.Context, c identity.Caller, act identity.Action) (map[string]struct{}, error) {
	accounts, err := a.lister.ActableAccounts(ctx, c, act)
	a.m.observe(err)
	return accounts, err //nolint:wrapcheck // the wrapped authorizer's error, unchanged, is what callers match
}

// watch returns az with its authorizer wrapped so that m sees every outcome.
// The wrapper lists accounts exactly when the authorizer does, so the views
// keep their one-call list.
func (m *authzMonitor) watch(az authorization) authorization {
	base := monitored{inner: az.authorizer, m: m}
	if l, ok := az.authorizer.(identityapp.AccountLister); ok {
		az.authorizer = monitoredLister{monitored: base, lister: l}
	} else {
		az.authorizer = base
	}
	return az
}
