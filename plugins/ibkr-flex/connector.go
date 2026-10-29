package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/ingest/ibkrflex"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// institution names IBKR on every delivery.
const institution = "ibkr"

// The files of the instance's flex-token Secret (<instance>-flex-token, D-89), as the manifest
// mounts it at flexMount.
const (
	flexMount   = "/etc/openfief/ibkr-flex"
	tokenFile   = "token"
	queryIDFile = "query-id"
	// maxFileBytes caps an uploaded statement, as the Flex parser does.
	maxFileBytes = 64 << 20
)

// errNoCredentials reports a credentials file that is missing or empty.
var errNoCredentials = errors.New("no Flex Web Service credentials")

// statementSource fetches a Flex statement from the Flex Web Service with a
// token and query ID.
type statementSource func(ctx context.Context, token, queryID string) ([]byte, error)

// connector maps IBKR Flex Query statements (internal/ingest/ibkrflex)
// through the connector contract: Sync pulls the statement of the
// configured Flex Query, ImportFile maps an uploaded one. The ingest mode
// comes from the request. The Flex Query sets the statement's period and
// re-importing it changes nothing (D-18), so Sync returns the watermark it
// was given.
type connector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	mapper  ibkrflex.Mapper
	fetch   statementSource
	secrets string
}

// Sync fetches and maps the Flex statement. The credentials are read from
// their files on every sync, so a rotated Secret is used at once; they are
// never logged or put in an error.
func (c *connector) Sync(ctx context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	mode, err := delivery.Mode(req.Msg.GetMode())
	if err != nil {
		return nil, status(connect.CodeInvalidArgument, err)
	}
	creds, err := c.credentials()
	if err != nil {
		return nil, status(connect.CodeFailedPrecondition, err)
	}
	statement, err := c.fetch(ctx, creds.token, creds.queryID)
	if err != nil {
		return nil, status(fetchCode(err), fmt.Errorf("ibkr-flex: %w", err))
	}
	d, err := c.mapStatement(req.Msg.GetBatchId(), statement, mode)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&connectorv1.SyncResponse{Delivery: d, Watermark: req.Msg.GetWatermark()}), nil
}

// ImportFile maps an uploaded Flex statement in the header's mode.
func (c *connector) ImportFile(_ context.Context, stream *connect.ClientStream[connectorv1.ImportFileRequest]) (*connect.Response[connectorv1.ImportFileResponse], error) {
	header, data, err := openfief.ReceiveFile(stream, maxFileBytes)
	if err != nil {
		return nil, status(connect.CodeInvalidArgument, err)
	}
	mode, err := delivery.Mode(header.GetMode())
	if err != nil {
		return nil, status(connect.CodeInvalidArgument, err)
	}
	d, err := c.mapStatement(header.GetBatchId(), data, mode)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&connectorv1.ImportFileResponse{Delivery: d}), nil
}

// mapStatement maps a statement into the delivery for batch. A statement
// that can't be read is InvalidArgument; rows it can't book are ingest
// problems in the delivery.
func (c *connector) mapStatement(batch string, statement []byte, mode ingest.Mode) (*connectorv1.Delivery, error) {
	m, err := c.mapper.Map(bytes.NewReader(statement), mode)
	if err != nil {
		return nil, status(connect.CodeInvalidArgument, fmt.Errorf("ibkr-flex: %w", err))
	}
	d, err := delivery.FromMapping(batch, institution, &m)
	if err != nil {
		return nil, status(connect.CodeInternal, fmt.Errorf("ibkr-flex: %w", err))
	}
	return d, nil
}

// status is the Connect error with code for err.
func status(code connect.Code, err error) *connect.Error { return connect.NewError(code, err) }

// flexCredentials are what the Flex Web Service takes: the token, which is
// secret, and the Flex Query's ID.
type flexCredentials struct {
	token   string
	queryID string
}

// credentials reads the token and query ID from the mounted Secret.
func (c *connector) credentials() (flexCredentials, error) {
	token, err := readSecretFile(c.secrets, tokenFile)
	if err != nil {
		return flexCredentials{}, err
	}
	queryID, err := readSecretFile(c.secrets, queryIDFile)
	if err != nil {
		return flexCredentials{}, err
	}
	return flexCredentials{token: token, queryID: queryID}, nil
}

// readSecretFile reads one file of the Secret in dir, through an os.Root so
// the read can't leave dir. An error names the file, never its content.
func readSecretFile(dir, name string) (string, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return "", fmt.Errorf("%w: %s", errNoCredentials, dir)
	}
	b, readErr := root.ReadFile(name)
	closeErr := root.Close()
	v := strings.TrimSpace(string(b))
	if readErr != nil || closeErr != nil || v == "" {
		return "", fmt.Errorf("%w: %s", errNoCredentials, filepath.Join(dir, name))
	}
	return v, nil
}

// fetchCode is the Connect code of a failed fetch: IBKR not answering, or
// not ready yet, is Unavailable, so the core retries the batch; an
// oversized statement is InvalidArgument.
func fetchCode(err error) connect.Code {
	if errors.Is(err, ibkrflex.ErrTooLarge) {
		return connect.CodeInvalidArgument
	}
	return connect.CodeUnavailable
}
