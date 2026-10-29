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

// ChunkSize is how many bytes of a file one ImportFile message carries.
const ChunkSize = 64 << 10

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

// Sync asks instance for its data since watermark as batch id.
func (c *Client) Sync(ctx context.Context, instance string, id ingest.BatchID, mode ingest.Mode, watermark string) (app.Delivery, error) {
	call, err := c.service(ctx, instance)
	if err != nil {
		return app.Delivery{}, err
	}
	m, err := modeTo(mode)
	if err != nil {
		return app.Delivery{}, err
	}
	req := connect.NewRequest(&connectorv1.SyncRequest{BatchId: string(id), Mode: m, Watermark: watermark})
	req.Header().Set("Authorization", call.auth)
	resp, err := call.svc.Sync(ctx, req)
	if err != nil {
		return app.Delivery{}, fmt.Errorf("connector %s: %w", instance, err)
	}
	return deliveryFrom(resp.Msg.GetDelivery(), id, resp.Msg.GetWatermark())
}

// ImportFile streams file to instance as batch id: a header, then the
// contents in chunks of ChunkSize bytes.
func (c *Client) ImportFile(ctx context.Context, instance string, id ingest.BatchID, mode ingest.Mode, file app.NamedFile) (app.Delivery, error) {
	call, err := c.service(ctx, instance)
	if err != nil {
		return app.Delivery{}, err
	}
	m, err := modeTo(mode)
	if err != nil {
		return app.Delivery{}, err
	}
	stream := call.svc.ImportFile(ctx)
	stream.RequestHeader().Set("Authorization", call.auth)
	header := &connectorv1.FileHeader{BatchId: string(id), Mode: m, FileName: file.Name, MediaType: file.MediaType}
	sendErr := stream.Send(&connectorv1.ImportFileRequest{Part: &connectorv1.ImportFileRequest_Header{Header: header}})
	if sendErr == nil {
		sendErr = sendChunks(stream, file.Contents)
	}
	resp, err := stream.CloseAndReceive()
	if err != nil {
		return app.Delivery{}, fmt.Errorf("connector %s: %w", instance, errors.Join(err, ignoreEOF(sendErr)))
	}
	if sendErr != nil {
		return app.Delivery{}, fmt.Errorf("connector %s: %w", instance, sendErr)
	}
	return deliveryFrom(resp.Msg.GetDelivery(), id, "")
}

// ImportBondLots hands lots to instance as batch id. An INVALID_ARGUMENT
// answer wraps app.ErrEntryRefused with the connector's reason, an
// UNIMPLEMENTED one app.ErrNoEntries.
func (c *Client) ImportBondLots(ctx context.Context, instance string, id ingest.BatchID, lots []ingest.BondLot) (app.Delivery, error) {
	call, err := c.service(ctx, instance)
	if err != nil {
		return app.Delivery{}, fmt.Errorf("%w: %w", app.ErrConnectorUnavailable, err)
	}
	req := connect.NewRequest(&connectorv1.ImportBondLotsRequest{BatchId: string(id), Lots: bondLotsTo(lots)})
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
// UNAVAILABLE.
func entryError(err error) error {
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

// sendChunks sends src's bytes on stream in chunks of ChunkSize. A send that
// fails with io.EOF means the connector answered early; CloseAndReceive
// returns its answer.
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
			return fmt.Errorf("read file: %w", err)
		}
	}
}

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
	return call{svc: connectorv1connect.NewConnectorServiceClient(c.http, url), auth: "Bearer " + token}, nil
}
