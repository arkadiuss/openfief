package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// relayBatch is how many events Relay.Publish reads at a time.
const relayBatch = 100

// Relay hands the outbox's events to a publisher, in order (D-130).
type Relay struct {
	Outbox    Outbox
	Publisher Publisher
	Now       func() time.Time
}

// Publish publishes every unpublished event, oldest first, marking each
// published after its publisher took it, and returns how many it
// published. It stops at the first failure, so an event is never skipped:
// the next Publish starts with it again.
func (r *Relay) Publish(ctx context.Context) (int, error) {
	n := 0
	for {
		events, err := r.Outbox.Unpublished(ctx, relayBatch)
		if err != nil {
			return n, fmt.Errorf("reading the outbox: %w", err)
		}
		for i := range events {
			if err := r.Publisher.Publish(ctx, &events[i]); err != nil {
				return n, fmt.Errorf("publishing event %d: %w", events[i].ID, err)
			}
			if err := r.Outbox.MarkPublished(ctx, events[i].ID, r.Now().UTC()); err != nil {
				return n, fmt.Errorf("marking event %d published: %w", events[i].ID, err)
			}
			n++
		}
		if len(events) < relayBatch {
			return n, nil
		}
	}
}

// LogPublisher publishes events to a log, the publisher until
// notifications exist. It logs internal IDs, the kind and the change, never
// the subject or detail, which may name a broker account or an amount
// (GR-12).
type LogPublisher struct {
	Log *slog.Logger
}

// Publish logs e.
func (p LogPublisher) Publish(ctx context.Context, e *Event) error {
	p.Log.LogAttrs(ctx, slog.LevelInfo, "attention item changed",
		slog.Int64("event", e.ID), slog.Int64("item", e.ItemID), slog.String("kind", string(e.Kind)),
		slog.Int("seq", e.Change.Seq), slog.String("state", e.Change.State.String()),
		slog.String("reason", e.Change.Reason.String()))
	return nil
}
