package main

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation/app"
	"github.com/arkadiuss/openfief/internal/valuation/nbp"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
)

// institution names NBP on every delivery.
const institution = "nbp"

// errBadWatermark reports a watermark this connector never returned.
var errBadWatermark = errors.New("watermark is not a date")

// connector serves NBP table A FX rates (D-20) through the connector
// contract. Its watermark is the date of the latest rates it delivered, so
// each sync asks only for the days after it, up to today. A first sync,
// without a watermark, starts at start.
type connector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	rates app.RateProvider
	today func() kernel.Date
	start kernel.Date
}

// Sync returns table A's rates from the day after the watermark to today.
// When NBP has published nothing new, such as on a weekend, the delivery is
// empty and the watermark stays. A failed request to NBP is Unavailable, so
// the core retries the batch.
func (c *connector) Sync(ctx context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	from, err := c.from(req.Msg.GetWatermark())
	if err != nil {
		return nil, status(connect.CodeInvalidArgument, err)
	}
	watermark := req.Msg.GetWatermark()
	out := &connectorv1.Delivery{BatchId: req.Msg.GetBatchId(), Institution: institution}
	if to := c.today(); !from.After(to) {
		rates, err := app.FetchRates(ctx, c.rates, from, to, nbp.MaxRangeDays)
		if err != nil {
			return nil, status(connect.CodeUnavailable, fmt.Errorf("nbp-fx: %w", err))
		}
		out.FxRates = delivery.FXRates(rates)
		if len(rates) > 0 {
			watermark = rates[len(rates)-1].Date.String() // FetchRates sorts by date
		}
	}
	return connect.NewResponse(&connectorv1.SyncResponse{Delivery: out, Watermark: watermark}), nil
}

// status is the Connect error with code for err.
func status(code connect.Code, err error) *connect.Error { return connect.NewError(code, err) }

// from is the first day a sync with watermark asks for.
func (c *connector) from(watermark string) (kernel.Date, error) {
	if watermark == "" {
		return c.start, nil
	}
	last, err := kernel.ParseDate(watermark)
	if err != nil {
		return kernel.Date{}, fmt.Errorf("%w: %q", errBadWatermark, watermark)
	}
	return last.AddDays(1), nil
}
