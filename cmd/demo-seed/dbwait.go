package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// errDatabaseNotReady reports a database that still refused connections when
// the loader stopped waiting for it.
var errDatabaseNotReady = errors.New("the database did not accept connections in time")

// Backoff of the database wait: the first retry after dbWaitFirst, each next
// one twice as late, never later than dbWaitMax.
const (
	dbWaitFirst = 500 * time.Millisecond
	dbWaitMax   = 5 * time.Second
	// dbWaitGrowth is how much longer each next retry waits.
	dbWaitGrowth = 2
)

// waitForDatabase calls ping until it succeeds, with a backoff between tries,
// for at most limit, so a Job started with the database does not spend its
// attempts on a server that is still starting. sleep waits for a duration
// or until ctx ends; the wait counts the time it slept. A ping that
// still fails at the limit is errDatabaseNotReady, wrapping the last error.
func waitForDatabase(ctx context.Context, ping func(context.Context) error, sleep func(context.Context, time.Duration) error, limit time.Duration) error {
	delay, waited := dbWaitFirst, time.Duration(0)
	for {
		err := ping(ctx)
		if err == nil {
			return nil
		}
		if waited >= limit {
			return fmt.Errorf("%w after %s: %w", errDatabaseNotReady, limit, err)
		}
		d := min(delay, limit-waited)
		if sleepErr := sleep(ctx, d); sleepErr != nil {
			return fmt.Errorf("waiting for the database: %w", sleepErr)
		}
		waited += d
		delay = min(dbWaitGrowth*delay, dbWaitMax)
	}
}

// sleepCtx waits for d or until ctx ends.
func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("waiting: %w", ctx.Err())
	case <-t.C:
		return nil
	}
}
