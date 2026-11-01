package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/attention"
)

// AttentionItems reads the attention items (D-130);
// *attention/app.Items implements it.
type AttentionItems interface {
	// List returns every item, in ID order, with its history.
	List(ctx context.Context) ([]attention.Item, error)
}

// WithAttention lists the attention items from items; without it the read
// API lists none.
func WithAttention(items AttentionItems) Option {
	return func(v *Views) { v.attention = items }
}

// Attention returns every attention item, in ID order, with its history,
// and how many are open.
func (v *Views) Attention(ctx context.Context) ([]attention.Item, int, error) {
	if v.attention == nil {
		return nil, 0, nil
	}
	items, err := v.attention.List(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("loading attention items: %w", err)
	}
	open := 0
	for i := range items {
		if items[i].State() == attention.StateOpen {
			open++
		}
	}
	return items, open, nil
}
