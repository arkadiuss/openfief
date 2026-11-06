package connectapi

import (
	"context"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// pageSize returns a request's page size, DefaultPageSize for 0.
func pageSize(size int32) int {
	if size == 0 {
		return views.DefaultPageSize
	}
	return int(size)
}

// ListActivity answers GET /v1/activity: the recent activity, newest first,
// a page at a time (D-193).
func (s *readService) ListActivity(ctx context.Context, req *connect.Request[viewsv1.ListActivityRequest]) (*connect.Response[viewsv1.ListActivityResponse], error) {
	var after *views.ActivityCursor
	if token := req.Msg.GetPageToken(); token != "" {
		c, err := views.ParseActivityCursor(token)
		if err != nil {
			return nil, toConnectError(ctx, s.log, err)
		}
		after = &c
	}
	page, err := s.views.Activity(ctx, after, pageSize(req.Msg.GetPageSize()))
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListActivityResponse{Activities: make([]*viewsv1.Activity, 0, len(page.Activities))}
	for i := range page.Activities {
		out.Activities = append(out.Activities, pbActivity(&page.Activities[i]))
	}
	if page.Next != nil {
		out.NextPageToken = page.Next.Encode()
	}
	return connect.NewResponse(out), nil
}

// pbActivity converts one activity.
func pbActivity(a *views.Activity) *viewsv1.Activity {
	out := &viewsv1.Activity{Kind: viewsv1.ActivityKind(enumNumber(int(a.Kind))), At: pbTime(a.At)}
	switch {
	case a.Batch != nil:
		b := a.Batch
		out.ConnectorInstance = b.Instance
		out.Batch = &viewsv1.ActivityBatch{
			BatchId: string(b.ID), Source: b.Source, Mode: b.Mode.String(), Institution: string(b.Institution),
			Transactions: clampInt32(b.Transactions), NewTransactions: clampInt32(b.New), Corrected: clampInt32(b.NewVersions),
			FailedAssertions: clampInt32(b.FailedAssertions), Problems: clampInt32(b.Problems), Unsupported: clampInt32(b.Unsupported),
		}
		for _, acc := range b.Accounts {
			out.Batch.AccountIds = append(out.Batch.AccountIds, string(acc))
		}
		out.Batch.Accounts = pbBatchAccounts(b.PerAccount)
	case a.Attempt != nil:
		out.ConnectorInstance = a.Attempt.Instance
		out.Attempt = pbSyncAttempt(a.Attempt)
	case a.Change != nil:
		out.Change = pbChange(a.Change)
	}
	return out
}

// pbBatchAccounts converts a batch's counts per account (D-199).
func pbBatchAccounts(counts []ingest.AccountCounts) []*viewsv1.BatchAccount {
	out := make([]*viewsv1.BatchAccount, 0, len(counts))
	for _, c := range counts {
		out = append(out, &viewsv1.BatchAccount{
			AccountId: string(c.Account), Transactions: clampInt32(c.Transactions), NewTransactions: clampInt32(c.New),
			Corrected: clampInt32(c.NewVersions), FailedAssertions: clampInt32(c.FailedAssertions),
		})
	}
	return out
}

// pbChange converts a coupon or a correction: a deletion has no kind; a
// coupon names its instrument and the cash it brought.
func pbChange(v *ledger.VersionedTransaction) *viewsv1.ActivityChange {
	t := &v.Transaction
	out := &viewsv1.ActivityChange{
		AccountId: string(t.Account), Source: t.Key.Source, ExternalId: t.Key.ExternalID, Version: clampInt32(v.Version),
		Date: pbDate(t.Date),
	}
	if t.Kind == ledger.KindDeletion {
		out.Deleted = true
		return out
	}
	out.Kind = viewsv1.TransactionKind(enumNumber(int(t.Kind)))
	if v.Version != 1 || t.Kind != ledger.KindInterest {
		return out
	}
	for i := range t.Postings {
		switch p := &t.Postings[i]; p.Book.Kind {
		case ledger.BookHolding:
			out.Instrument = pbInstrument(p.Book.Instrument)
		case ledger.BookCash:
			out.Amount = pbMoney(p.Amount)
		case ledger.BookFee, ledger.BookWithholdingTax, ledger.BookFX, ledger.BookEquity:
		}
	}
	return out
}

// pbTime writes t in RFC 3339, UTC; "" for the zero time.
func pbTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// pbSyncAttempt converts a sync attempt.
func pbSyncAttempt(a *ingest.SyncAttempt) *viewsv1.SyncAttempt {
	return &viewsv1.SyncAttempt{
		ConnectorInstance: a.Instance, BatchId: string(a.Batch), Mode: a.Mode.String(),
		StartedAt: pbTime(a.Started), EndedAt: pbTime(a.Ended), Failed: a.Failed(),
		ErrorClass: viewsv1.SyncErrorClass(enumNumber(int(a.Class))), Failures: clampInt32(a.Failures),
		NextSyncAt: pbTime(a.NextSync),
	}
}

// pbSyncStatus converts an instance's sync status.
func pbSyncStatus(st *ingest.SyncStatus) *viewsv1.SyncStatus {
	return &viewsv1.SyncStatus{
		ConnectorInstance: st.Instance, LastAttemptAt: pbTime(st.LastAttempt), LastSuccessAt: pbTime(st.LastSuccess),
		Failures: clampInt32(st.Failures), FailingSince: pbTime(st.FailingSince),
		LastErrorClass: viewsv1.SyncErrorClass(enumNumber(int(st.LastClass))), NextSyncAt: pbTime(st.NextSync),
		KeepsFailing: st.KeepsFailing(),
	}
}

// ListSyncStatuses answers GET /v1/sync-status (D-196).
func (s *readService) ListSyncStatuses(ctx context.Context, _ *connect.Request[viewsv1.ListSyncStatusesRequest]) (*connect.Response[viewsv1.ListSyncStatusesResponse], error) {
	statuses, err := s.views.SyncStatuses(ctx)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListSyncStatusesResponse{Statuses: make([]*viewsv1.SyncStatus, 0, len(statuses))}
	for i := range statuses {
		out.Statuses = append(out.Statuses, pbSyncStatus(&statuses[i]))
	}
	return connect.NewResponse(out), nil
}

// ListSyncAttempts answers GET
// /v1/connector-instances/{connector_instance}/sync-attempts (D-196).
func (s *readService) ListSyncAttempts(ctx context.Context, req *connect.Request[viewsv1.ListSyncAttemptsRequest]) (*connect.Response[viewsv1.ListSyncAttemptsResponse], error) {
	var after views.ActivityBound
	if token := req.Msg.GetPageToken(); token != "" {
		b, err := views.ParseActivityBound(token)
		if err != nil {
			return nil, toConnectError(ctx, s.log, err)
		}
		after = b
	}
	page, err := s.views.SyncAttempts(ctx, req.Msg.GetConnectorInstance(), after, pageSize(req.Msg.GetPageSize()))
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListSyncAttemptsResponse{Status: pbSyncStatus(&page.Status), Attempts: make([]*viewsv1.SyncAttempt, 0, len(page.Attempts))}
	for i := range page.Attempts {
		a := pbSyncAttempt(&page.Attempts[i])
		a.Accounts = pbBatchAccounts(page.Booked[page.Attempts[i].Batch])
		out.Attempts = append(out.Attempts, a)
	}
	if page.Next != nil {
		out.NextPageToken = page.Next.Encode()
	}
	return connect.NewResponse(out), nil
}
