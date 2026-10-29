package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/kernel"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
)

// connector maps treasury bond lots the owner entered
// (internal/ingest/treasurybonds) through the connector contract. It has no
// institution to pull from, so a Sync delivers nothing and keeps its
// watermark; it takes no files.
type connector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	mapper treasurybonds.Mapper
	today  func() kernel.Date
}

// Sync delivers nothing: every lot arrives through ImportBondLots. It
// answers so that the core's scheduled syncs of the instance succeed.
func (c *connector) Sync(_ context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	d := &connectorv1.Delivery{BatchId: req.Msg.GetBatchId(), Institution: string(treasurybonds.Institution)}
	return connect.NewResponse(&connectorv1.SyncResponse{Delivery: d, Watermark: req.Msg.GetWatermark()}), nil
}

// status is the Connect error with code for err.
func status(code connect.Code, err error) *connect.Error { return connect.NewError(code, err) }

// ImportBondLots maps the lots into deposits and buys in the configured
// account, plus each series' terms. A lot that doesn't parse or breaks a
// rule is InvalidArgument, naming the rule, and nothing is delivered.
func (c *connector) ImportBondLots(_ context.Context, req *connect.Request[connectorv1.ImportBondLotsRequest]) (*connect.Response[connectorv1.ImportBondLotsResponse], error) {
	lots, err := delivery.BondLots(req.Msg.GetLots())
	if err != nil {
		return nil, status(connect.CodeInvalidArgument, err)
	}
	m, terms, err := c.mapper.Map(lots, c.today())
	if err != nil {
		return nil, status(connect.CodeInvalidArgument, err)
	}
	d, err := delivery.FromMapping(req.Msg.GetBatchId(), string(treasurybonds.Institution), &m)
	if err != nil {
		return nil, status(connect.CodeInternal, fmt.Errorf("treasury-bonds: %w", err))
	}
	d.BondTerms = delivery.BondTerms(terms)
	return connect.NewResponse(&connectorv1.ImportBondLotsResponse{Delivery: d}), nil
}
