package main

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/ingest/xtb"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// institution names XTB on every delivery.
const institution = "xtb"

// connector maps uploaded xStation exports (internal/ingest/xtb) through the
// connector contract. XTB has no API it may reach, so a sync delivers
// nothing and keeps its watermark.
type connector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	mapper xtb.Mapper
}

// Sync delivers nothing: an XTB account's statements come only as uploads.
func (c *connector) Sync(_ context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	d := &connectorv1.Delivery{BatchId: req.Msg.GetBatchId(), Institution: institution}
	return connect.NewResponse(&connectorv1.SyncResponse{Delivery: d, Watermark: req.Msg.GetWatermark()}), nil
}

// ImportFile maps an uploaded export in the header's mode. A file that
// isn't a readable export, or an opening batch, is InvalidArgument with the
// reason, which never quotes the file.
func (c *connector) ImportFile(_ context.Context, stream *connect.ClientStream[connectorv1.ImportFileRequest]) (*connect.Response[connectorv1.ImportFileResponse], error) {
	header, data, err := openfief.ReceiveFile(stream, xtb.MaxFileBytes)
	if err != nil {
		return nil, status(connect.CodeInvalidArgument, err)
	}
	mode, err := delivery.Mode(header.GetMode())
	if err != nil {
		return nil, status(connect.CodeInvalidArgument, err)
	}
	m, err := c.mapper.Map(data, header.GetFileName(), mode)
	if err != nil {
		return nil, status(mapCode(err), fmt.Errorf("xtb: %w", err))
	}
	d, err := delivery.FromMapping(header.GetBatchId(), institution, &m)
	if err != nil {
		return nil, status(connect.CodeInternal, fmt.Errorf("xtb: %w", err))
	}
	return connect.NewResponse(&connectorv1.ImportFileResponse{Delivery: d}), nil
}

// mapCode is the Connect code of a file the mapper refused: the owner's
// file or mode is wrong, so InvalidArgument; anything else is Internal.
func mapCode(err error) connect.Code {
	for _, e := range []error{xtb.ErrNotWorkbook, xtb.ErrTooLarge, xtb.ErrLayout, xtb.ErrCurrency, xtb.ErrOpeningMode, xtb.ErrUnknownMode} {
		if errors.Is(err, e) {
			return connect.CodeInvalidArgument
		}
	}
	return connect.CodeInternal
}

// status is the Connect error with code for err.
func status(code connect.Code, err error) *connect.Error { return connect.NewError(code, err) }
