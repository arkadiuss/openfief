// Package connectrpc calls connector plugins' connector contract,
// openfief.connector.v1.ConnectorService, over Connect RPC (D-11, D-49): the
// core's side of a plugin connector's ingest path (D-7). It authenticates
// each call as the core with a token whose audience is the instance (D-59),
// and translates the contract's messages into the domain, so they never
// reach it.
package connectrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
)

const (
	// ChunkSize is how many bytes of a file one ImportFile message carries.
	ChunkSize = 64 << 10
	// MaxDeliveryBytes caps a connector's answer (D-184): a market-data
	// backfill of MaxDeliveryItems prices, about 10 MiB, fits within it. A
	// larger answer fails the call with RESOURCE_EXHAUSTED before it is
	// decoded. With MaxDeliveryMessages it keeps the worst answer the core
	// decodes and translates to about a third of its 256 MiB memory limit.
	MaxDeliveryBytes = 16 << 20
)

// Client is app.ConnectorCaller over Connect RPC.
type Client struct {
	http    connect.HTTPClient
	baseURL func(instance string) (string, error)
	tokens  app.CoreTokens
}

// NewClient returns a client that reaches an instance at the URL baseURL
// returns for it, such as its ClusterIP Service, with httpClient, and
// authenticates with a token from tokens. ImportFile streams, which needs
// HTTP/2: H2CClient returns a client that speaks it without TLS.
func NewClient(httpClient connect.HTTPClient, baseURL func(instance string) (string, error), tokens app.CoreTokens) *Client {
	return &Client{http: httpClient, baseURL: baseURL, tokens: tokens}
}

// H2CClient returns an HTTP client that speaks HTTP/2 without TLS, as the
// in-cluster calls to plugins do.
func H2CClient() *http.Client {
	var p http.Protocols
	p.SetUnencryptedHTTP2(true)
	return &http.Client{Transport: &http.Transport{Protocols: &p}}
}

// Sync asks instance for its data since the call's watermark as the call's
// batch, with what a market-data connector is to price (D-93), what the
// instance's bound accounts hold (D-117) and the instruments they post to
// under a provisional ISIN (D-214).
func (c *Client) Sync(ctx context.Context, instance string, sc *app.SyncCall) (app.Delivery, error) {
	call, err := c.service(ctx, instance)
	if err != nil {
		return app.Delivery{}, fmt.Errorf("%w: %w", app.ErrConnectorUnavailable, err)
	}
	m, err := modeTo(sc.Mode)
	if err != nil {
		return app.Delivery{}, err
	}
	req := connect.NewRequest(&connectorv1.SyncRequest{
		BatchId: string(sc.ID), Mode: m, Watermark: sc.Watermark,
		Instruments: instrumentsTo(sc.Scope.Instruments), Indices: sc.Scope.Indices,
		Accounts: heldTo(sc.Held), BondTerms: bondTermsTo(sc.BondTerms),
		ProvisionalInstruments: instrumentsTo(sc.Provisional),
	})
	req.Header().Set("Authorization", call.auth)
	resp, err := call.svc.Sync(ctx, req)
	if err != nil {
		return app.Delivery{}, fmt.Errorf("connector %s: %w", instance, syncError(err))
	}
	return deliveryFrom(resp.Msg.GetDelivery(), sc.ID, resp.Msg.GetWatermark())
}

// ImportFile streams file to instance as batch id: a header, with the
// opening cash the owner entered when there is one (D-136), then the
// contents in chunks of ChunkSize bytes. A file that can't be read stops
// the call, so the connector maps nothing, and the read's error is
// returned. The connector's INVALID_ARGUMENT wraps app.ErrEntryRefused with
// its reason, its UNIMPLEMENTED app.ErrNoEntries, and an instance that
// can't be reached app.ErrConnectorUnavailable.
func (c *Client) ImportFile(ctx context.Context, instance string, id ingest.BatchID, mode ingest.Mode, file app.NamedFile) (app.Delivery, error) {
	call, err := c.service(ctx, instance)
	if err != nil {
		return app.Delivery{}, fmt.Errorf("%w: %w", app.ErrConnectorUnavailable, err)
	}
	m, err := modeTo(mode)
	if err != nil {
		return app.Delivery{}, err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream := call.svc.ImportFile(ctx)
	stream.RequestHeader().Set("Authorization", call.auth)
	header := &connectorv1.FileHeader{BatchId: string(id), Mode: m, FileName: file.Name, MediaType: file.MediaType}
	if file.OpeningCash != nil {
		header.OpeningCash = moneyTo(*file.OpeningCash)
	}
	sendErr := stream.Send(&connectorv1.ImportFileRequest{Part: &connectorv1.ImportFileRequest_Header{Header: header}})
	if sendErr == nil {
		sendErr = sendChunks(stream, file.Contents)
		if readErr, unread := errors.AsType[*readError](sendErr); unread {
			// Cancel without closing the stream: closing it would end the
			// file cleanly, and the connector could map the truncated file.
			cancel()
			return app.Delivery{}, fmt.Errorf("connector %s: %w", instance, readErr.err)
		}
	}
	resp, err := stream.CloseAndReceive()
	if err != nil {
		return app.Delivery{}, fmt.Errorf("connector %s: %w", instance, errors.Join(entryError(err), ignoreEOF(sendErr)))
	}
	if sendErr != nil {
		return app.Delivery{}, fmt.Errorf("connector %s: %w", instance, sendErr)
	}
	return deliveryFrom(resp.Msg.GetDelivery(), id, "")
}

// ImportBondLots hands e's account, lots, deleted lots and go-live date, if any, to instance
// as batch id. An INVALID_ARGUMENT answer wraps app.ErrEntryRefused with
// the connector's reason, an UNIMPLEMENTED one app.ErrNoEntries.
func (c *Client) ImportBondLots(ctx context.Context, instance string, id ingest.BatchID, e *ingest.BondLotEntries) (app.Delivery, error) {
	call, err := c.service(ctx, instance)
	if err != nil {
		return app.Delivery{}, fmt.Errorf("%w: %w", app.ErrConnectorUnavailable, err)
	}
	req := connect.NewRequest(&connectorv1.ImportBondLotsRequest{
		BatchId: string(id), Lots: bondLotsTo(e.Lots), Deleted: bondLotKeysTo(e.Deleted), AccountId: string(e.Account),
		Booked: bookedTo(e.Booked),
	})
	if !e.GoLive.IsZero() {
		req.Msg.GoLive = e.GoLive.String()
	}
	req.Header().Set("Authorization", call.auth)
	resp, err := call.svc.ImportBondLots(ctx, req)
	if err != nil {
		return app.Delivery{}, fmt.Errorf("connector %s: %w", instance, entryError(err))
	}
	return deliveryFrom(resp.Msg.GetDelivery(), id, "")
}

// entryError returns err wrapping app.ErrEntryRefused, with the connector's
// own message, when the connector answered INVALID_ARGUMENT; wrapping
// app.ErrNoEntries when it answered UNIMPLEMENTED; and wrapping
// app.ErrConnectorUnavailable when it could not be reached or answered
// UNAVAILABLE. An answer the core refused to decode as too large stays
// ErrDeliveryTooLarge.
func entryError(err error) error {
	if errors.Is(err, ErrDeliveryTooLarge) {
		return err
	}
	code := connect.CodeOf(err)
	if code == connect.CodeInvalidArgument {
		if ce, ok := errors.AsType[*connect.Error](err); ok {
			return fmt.Errorf("%w: %s", app.ErrEntryRefused, ce.Message())
		}
		return fmt.Errorf("%w: %w", app.ErrEntryRefused, err)
	}
	if code == connect.CodeUnimplemented {
		return fmt.Errorf("%w: %w", app.ErrNoEntries, err)
	}
	if code == connect.CodeUnavailable || code == connect.CodeDeadlineExceeded {
		return fmt.Errorf("%w: %w", app.ErrConnectorUnavailable, err)
	}
	return err
}

// syncError returns err, a failed Sync call, wrapping what the scheduler
// classifies it by (app.SyncErrorClassOf), never by its text: a context's
// end as it is; an answer too large to read app.ErrInvalidDelivery; an
// instance that could not be reached or answered UNAVAILABLE or
// DEADLINE_EXCEEDED app.ErrConnectorUnavailable; and any other answer of
// the connector app.ErrConnectorFailed.
func syncError(err error) error {
	switch code := connect.CodeOf(err); {
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return err
	case errors.Is(err, ErrDeliveryTooLarge):
		return fmt.Errorf("%w: %w", app.ErrInvalidDelivery, err)
	case code == connect.CodeUnavailable, code == connect.CodeDeadlineExceeded:
		return fmt.Errorf("%w: %w", app.ErrConnectorUnavailable, err)
	}
	return fmt.Errorf("%w: %w", app.ErrConnectorFailed, err)
}

// sendChunks sends src's bytes on stream in chunks of ChunkSize. A send that
// fails with io.EOF means the connector answered early; CloseAndReceive
// returns its answer. A failed read is a *readError.
func sendChunks(stream *connect.ClientStreamForClient[connectorv1.ImportFileRequest, connectorv1.ImportFileResponse], src io.Reader) error {
	buf := make([]byte, ChunkSize)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			chunk := &connectorv1.ImportFileRequest{Part: &connectorv1.ImportFileRequest_Chunk{Chunk: append([]byte(nil), buf[:n]...)}}
			if sendErr := stream.Send(chunk); sendErr != nil {
				return fmt.Errorf("send file: %w", sendErr)
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return &readError{err: fmt.Errorf("read file: %w", err)}
		}
	}
}

// readError is a file that could not be read while it was sent.
type readError struct {
	err error
}

// Error returns the read's error.
func (r *readError) Error() string { return r.err.Error() }

// Unwrap returns the read's error.
func (r *readError) Unwrap() error { return r.err }

// ignoreEOF drops a send error that only says the stream was closed.
func ignoreEOF(err error) error {
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

// call is one authenticated call to an instance: its generated client and
// the Authorization header value with a fresh core token.
type call struct {
	svc  connectorv1connect.ConnectorServiceClient
	auth string
}

// service returns the call for instance.
func (c *Client) service(ctx context.Context, instance string) (call, error) {
	url, err := c.baseURL(instance)
	if err != nil {
		return call{}, fmt.Errorf("connector %s address: %w", instance, err)
	}
	token, err := c.tokens.TokenFor(ctx, instance)
	if err != nil {
		return call{}, fmt.Errorf("core token for %s: %w", instance, err)
	}
	// The answer is read up to MaxDeliveryBytes, after decompression, and
	// its elements are counted on the wire before it is decoded, so a
	// connector can't make the core hold more (D-184).
	svc := connectorv1connect.NewConnectorServiceClient(c.http, url,
		connect.WithReadMaxBytes(MaxDeliveryBytes), connect.WithCodec(boundedCodec{}))
	return call{svc: svc, auth: "Bearer " + token}, nil
}
