package main

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/kernel"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
)

// connector maps treasury bond lots the owner entered
// (internal/ingest/treasurybonds) through the connector contract, and its
// Sync delivers the period index readings the bonds' later interest periods
// need: GUS's CPI and the NBP reference rate (D-9). It takes no files.
type connector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	mapper  treasurybonds.Mapper
	indices *treasurybonds.IndexSync
	coupons *treasurybonds.CouponSync
	today   func() kernel.Date
}

// Sync delivers the CPI and reference-rate readings for the months after
// the watermark, and the watermark of the last of each. In a transactions
// batch it also books, as interest, each coupon that fell due after go-live
// on the open lots the request says the bound accounts hold (D-101,
// D-117); an opening batch seeds nothing here. A watermark this connector
// never returned, or holdings that don't parse, is InvalidArgument; a
// failed request to GUS or NBP is Unavailable, so the core retries the
// batch.
func (c *connector) Sync(ctx context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	today := c.today()
	readings, watermark, err := c.indices.Readings(ctx, req.Msg.GetWatermark(), today)
	if errors.Is(err, treasurybonds.ErrBadWatermark) {
		return nil, status(connect.CodeInvalidArgument, err)
	}
	if err != nil {
		return nil, status(connect.CodeUnavailable, fmt.Errorf("treasury-bonds: %w", err))
	}
	msgs, err := delivery.IndexReadings(readings)
	if err != nil {
		return nil, status(connect.CodeInternal, fmt.Errorf("treasury-bonds: %w", err))
	}
	d := &connectorv1.Delivery{BatchId: req.Msg.GetBatchId(), Institution: string(treasurybonds.Institution), IndexReadings: msgs}
	if req.Msg.GetMode() == connectorv1.Mode_MODE_TRANSACTIONS {
		if d, err = withCoupons(ctx, c.coupons, req.Msg, d, today); err != nil {
			return nil, err
		}
	}
	return connect.NewResponse(&connectorv1.SyncResponse{Delivery: d, Watermark: watermark}), nil
}

// withCoupons returns d with the coupons, booked by coupons, due on what req says the bound
// accounts hold, as of today.
func withCoupons(ctx context.Context, coupons *treasurybonds.CouponSync, req *connectorv1.SyncRequest, d *connectorv1.Delivery, today kernel.Date) (*connectorv1.Delivery, error) {
	held, terms, err := delivery.Held(req)
	if err != nil {
		return nil, status(connect.CodeInvalidArgument, err)
	}
	m, err := coupons.Coupons(ctx, held, terms, today)
	if err != nil {
		return nil, status(connect.CodeUnavailable, fmt.Errorf("treasury-bonds: %w", err))
	}
	out, err := delivery.FromMapping(d.GetBatchId(), d.GetInstitution(), &m)
	if err != nil {
		return nil, status(connect.CodeInternal, fmt.Errorf("treasury-bonds: %w", err))
	}
	out.IndexReadings = d.GetIndexReadings()
	return out, nil
}

// status is the Connect error with code for err.
func status(code connect.Code, err error) *connect.Error { return connect.NewError(code, err) }

// ImportBondLots maps the lots into deposits and buys in the account the
// core names with them (D-118), plus each series' terms. A lot that doesn't parse or breaks a
// rule is InvalidArgument, naming the rule, and nothing is delivered.
func (c *connector) ImportBondLots(_ context.Context, req *connect.Request[connectorv1.ImportBondLotsRequest]) (*connect.Response[connectorv1.ImportBondLotsResponse], error) {
	entries, err := delivery.BondLotEntries(req.Msg)
	if err != nil {
		return nil, status(connect.CodeInvalidArgument, err)
	}
	m, terms, err := c.mapper.Map(&entries, c.today())
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
