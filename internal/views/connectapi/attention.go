package connectapi

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/views/app"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// ListAttentionItems answers GET /v1/attention-items: the attention items in
// the requested view, tied to the requested account, instrument or
// instance, each with its group, scope and history, and how many of those
// the filters select are open, acknowledged and done (D-130, D-196).
func (s *readService) ListAttentionItems(ctx context.Context, req *connect.Request[viewsv1.ListAttentionItemsRequest]) (*connect.Response[viewsv1.ListAttentionItemsResponse], error) {
	view, err := attentionView(req.Msg.GetView())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	q := app.AttentionQuery{
		View: view, Account: req.Msg.GetAccountId(), Instrument: req.Msg.GetIsin(), Instance: req.Msg.GetConnectorInstance(),
	}
	list, err := s.views.Attention(ctx, &q)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListAttentionItemsResponse{
		Items: make([]*viewsv1.AttentionItem, 0, len(list.Items)), Open: clampInt32(list.Open),
		Acknowledged: clampInt32(list.Acknowledged), Done: clampInt32(list.Done),
	}
	for i := range list.Items {
		out.Items = append(out.Items, pbAttentionItem(&list.Items[i]))
	}
	return connect.NewResponse(out), nil
}

// attentionView converts a requested view; unspecified means every item.
func attentionView(v viewsv1.AttentionView) (app.AttentionView, error) {
	switch v {
	case viewsv1.AttentionView_ATTENTION_VIEW_UNSPECIFIED:
		return 0, nil
	case viewsv1.AttentionView_ATTENTION_VIEW_OPEN:
		return app.AttentionOpen, nil
	case viewsv1.AttentionView_ATTENTION_VIEW_ACKNOWLEDGED:
		return app.AttentionAcknowledged, nil
	case viewsv1.AttentionView_ATTENTION_VIEW_DONE:
		return app.AttentionDone, nil
	}
	return 0, fmt.Errorf("%w: attention view %d", errBadParameter, v)
}

// pbAttentionItem converts an attention item with its history, group and
// scope, its subject as read access may see it (attention.ReadSubject,
// D-121, D-137).
func pbAttentionItem(it *app.ScopedItem) *viewsv1.AttentionItem {
	cur := it.Current()
	out := &viewsv1.AttentionItem{
		Id: it.ID, Kind: string(it.Kind), Subject: attention.ReadSubject(it.Kind, it.Subject), State: pbAttentionState(cur.State), Detail: cur.Detail,
		OwnerResolves: attention.OwnerResolves(it.Kind), History: make([]*viewsv1.AttentionItemChange, 0, len(it.Changes)),
		Group: viewsv1.AttentionItemGroup(enumNumber(int(it.Group))), AccountId: it.Scope.Account,
		Isin: it.Scope.Instrument, ConnectorInstance: it.Scope.Instance,
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
