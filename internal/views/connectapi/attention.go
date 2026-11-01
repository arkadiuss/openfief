package connectapi

import (
	"context"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/attention"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// ListAttentionItems answers GET /v1/attention-items: every attention item
// with its history, and how many are open (D-130).
func (s *readService) ListAttentionItems(ctx context.Context, _ *connect.Request[viewsv1.ListAttentionItemsRequest]) (*connect.Response[viewsv1.ListAttentionItemsResponse], error) {
	items, open, err := s.views.Attention(ctx)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListAttentionItemsResponse{Items: make([]*viewsv1.AttentionItem, 0, len(items)), Open: clampInt32(open)}
	for i := range items {
		out.Items = append(out.Items, pbAttentionItem(&items[i]))
	}
	return connect.NewResponse(out), nil
}

// pbAttentionItem converts an attention item with its history.
func pbAttentionItem(it *attention.Item) *viewsv1.AttentionItem {
	cur := it.Current()
	out := &viewsv1.AttentionItem{
		Id: it.ID, Kind: string(it.Kind), Subject: it.Subject, State: pbAttentionState(cur.State), Detail: cur.Detail,
		OwnerResolves: attention.OwnerResolves(it.Kind), History: make([]*viewsv1.AttentionItemChange, 0, len(it.Changes)),
	}
	for _, c := range it.Changes {
		out.History = append(out.History, &viewsv1.AttentionItemChange{
			Seq: clampInt32(c.Seq), State: pbAttentionState(c.State), Reason: pbAttentionReason(c.Reason),
			Detail: c.Detail, At: c.At.UTC().Format(time.RFC3339),
		})
	}
	return out
}

// pbAttentionState converts an item's state.
func pbAttentionState(s attention.State) viewsv1.AttentionItemState {
	switch s {
	case attention.StateOpen:
		return viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN
	case attention.StateAcknowledged:
		return viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_ACKNOWLEDGED
	case attention.StateResolved:
		return viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_RESOLVED
	}
	return viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_UNSPECIFIED
}

// pbAttentionReason converts why an item changed.
func pbAttentionReason(r attention.Reason) viewsv1.AttentionItemReason {
	switch r {
	case attention.ReasonRaised:
		return viewsv1.AttentionItemReason_ATTENTION_ITEM_REASON_RAISED
	case attention.ReasonChanged:
		return viewsv1.AttentionItemReason_ATTENTION_ITEM_REASON_CHANGED
	case attention.ReasonReopened:
		return viewsv1.AttentionItemReason_ATTENTION_ITEM_REASON_REOPENED
	case attention.ReasonCleared:
		return viewsv1.AttentionItemReason_ATTENTION_ITEM_REASON_CLEARED
	case attention.ReasonAcknowledged:
		return viewsv1.AttentionItemReason_ATTENTION_ITEM_REASON_ACKNOWLEDGED
	case attention.ReasonResolved:
		return viewsv1.AttentionItemReason_ATTENTION_ITEM_REASON_RESOLVED
	}
	return viewsv1.AttentionItemReason_ATTENTION_ITEM_REASON_UNSPECIFIED
}
