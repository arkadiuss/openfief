package main

import (
	"context"
	"fmt"
	"sync/atomic"

	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	ingestpg "github.com/arkadiuss/openfief/internal/ingest/postgres"
)

// syncTrigger is the connector scheduler the owner's sync retries go to
// (D-196): none until the core runs the scheduler, and then it. Until
// then, and outside a cluster, every retry fails with
// ingestapp.ErrSyncsNotRunning.
type syncTrigger struct {
	scheduler atomic.Pointer[ingestapp.Scheduler]
}

// Retry starts one sync of instance on the scheduler.
func (t *syncTrigger) Retry(ctx context.Context, instance string) error {
	s := t.scheduler.Load()
	if s == nil {
		return ingestapp.ErrSyncsNotRunning
	}
	if err := s.Retry(ctx, instance); err != nil {
		return fmt.Errorf("retrying a sync: %w", err)
	}
	return nil
}

// set makes s the scheduler retries go to.
func (t *syncTrigger) set(s *ingestapp.Scheduler) { t.scheduler.Store(s) }

// newRetries returns the owner's sync retries (D-196) over d: the owner's
// sessions and the authorizer checked by owners, the instance's bound
// accounts from the ingest schema, and the scheduler d.syncs holds once it
// runs, each instance at most once a minute.
func newRetries(d *connectorDeps, owners ingestapp.Owners) *ingestapp.Retries {
	retries := &ingestapp.Retries{
		Owners: owners, Bindings: ingestpg.New(d.pool), Trigger: d.syncs, Every: ingestapp.DefaultRetryEvery, Now: d.now,
	}
	if a, ok := owners.(ingestapp.ChangeAuditor); ok {
		retries.Audit = a
	}
	return retries
}
