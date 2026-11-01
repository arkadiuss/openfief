package main

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	attentiondomain "github.com/arkadiuss/openfief/internal/attention"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
)

// errAttentionDown is the fake attention items' failure.
var errAttentionDown = errors.New("attention store down")

// attentionCall is one call the fake attention items received.
type attentionCall struct {
	Op   string
	Cond attentiondomain.Condition
}

// fakeAttentionItems records the conditions raised and cleared, failing
// each call with err when it is set.
type fakeAttentionItems struct {
	calls []attentionCall
	err   error
}

// Raise records cond.
func (f *fakeAttentionItems) Raise(_ context.Context, cond attentiondomain.Condition) error {
	f.calls = append(f.calls, attentionCall{Op: "raise", Cond: cond})
	return f.err
}

// Clear records kind and subject.
func (f *fakeAttentionItems) Clear(_ context.Context, kind attentiondomain.Kind, subject string) error {
	f.calls = append(f.calls, attentionCall{Op: "clear", Cond: attentiondomain.Condition{Kind: kind, Subject: subject}})
	return f.err
}

// TestValuationAttention checks the adapter from valuation's attention port
// to the attention items (D-132): a raise becomes the condition with its
// kind, subject and detail, a clear names the kind and subject, and the
// attention store's failure reaches the revaluation.
func TestValuationAttention(t *testing.T) {
	t.Parallel()
	const isin = "PLTEST000010"
	tests := []struct {
		name    string
		clear   bool
		fail    error
		want    []attentionCall
		wantErr error
	}{
		{
			name: "raise",
			want: []attentionCall{{Op: "raise", Cond: attentiondomain.Condition{
				Kind: valuationapp.UnclassifiedKind, Subject: isin, Detail: "detail",
			}}},
		},
		{
			name:  "clear",
			clear: true,
			want:  []attentionCall{{Op: "clear", Cond: attentiondomain.Condition{Kind: valuationapp.UnclassifiedKind, Subject: isin}}},
		},
		{
			name: "raise fails",
			fail: errAttentionDown,
			want: []attentionCall{{Op: "raise", Cond: attentiondomain.Condition{
				Kind: valuationapp.UnclassifiedKind, Subject: isin, Detail: "detail",
			}}},
			wantErr: errAttentionDown,
		},
		{
			name:    "clear fails",
			clear:   true,
			fail:    errAttentionDown,
			want:    []attentionCall{{Op: "clear", Cond: attentiondomain.Condition{Kind: valuationapp.UnclassifiedKind, Subject: isin}}},
			wantErr: errAttentionDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			items := &fakeAttentionItems{err: tt.fail}
			var port valuationapp.AttentionRaiser = valuationAttention{items: items}
			var err error
			if tt.clear {
				err = port.Clear(t.Context(), valuationapp.UnclassifiedKind, isin)
			} else {
				err = port.Raise(t.Context(), valuationapp.UnclassifiedKind, isin, "detail")
			}
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, items.calls); diff != "" {
				t.Errorf("calls (-want +got):\n%s", diff)
			}
		})
	}
}
