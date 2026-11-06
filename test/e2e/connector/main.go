//go:build e2e

// Command connector is the end-to-end suite's synthetic connector plugin,
// built only for the suite (scripts/e2e-images.sh) and never released. It
// stands in for a broker connector that needs credentials, without calling
// any institution: its manifest (manifest.json) asks for the connector
// capability, the institution e2e-bank and one Secret role, credentials,
// mounted at /etc/openfief/e2e-connector; manifest-egress.json, the
// manifest of its second image, asks for the host example.org too, so the
// suite can upgrade one to the other (test/e2e/trust_test.go).
//
// Each sync reads the credential from the mounted Secret, as a real
// connector reads its token, and fails when it is missing, so a sync that
// succeeds shows the Secret reached this pod. It never logs, returns or
// delivers the credential. In transactions mode, with an account in its
// config, it delivers one synthetic deposit of 100.00 PLN into that account
// on the config's date (test/e2e/connector_test.go); in opening mode it
// delivers nothing, so the core never seeds an account with it.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"connectrpc.com/connect"

	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// The connector's synthetic delivery.
const (
	// institution is the institution its manifest declares.
	institution = "e2e-bank"
	// mountFile is the file of the Secret the manifest mounts at mountDir
	// that holds the credential.
	mountDir  = "/etc/openfief/e2e-connector"
	mountFile = "token"
	// depositSource and depositPrefix make the deposit's key.
	depositSource = "e2e-connector"
	depositPrefix = "deposit-"
	// depositAmount and depositCurrency are the deposit's amount.
	depositAmount   = "100.00"
	depositCurrency = "PLN"
)

// errNoCredential reports a credential file that is missing or empty.
var errNoCredential = errors.New("no credential in the mounted Secret")

// config is the instance's config: the account and date of its deposit.
type config struct {
	Account string `json:"account"`
	Date    string `json:"date"`
}

func main() {
	os.Exit(serve())
}

// serve runs the connector until SIGTERM or an interrupt and returns its
// exit code.
func serve() int {
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()
	var cfg config
	if err := openfief.LoadConfig(&cfg); err != nil {
		log.Error("e2e connector: config", "error", err.Error())
		return 1
	}
	log.Info("e2e connector serving")
	if err := openfief.ServeConnector(ctx, &connector{cfg: cfg, credentials: mountDir}); err != nil {
		log.Error("e2e connector stopped", "error", err.Error())
		return 1
	}
	return 0
}

// connector answers Sync with the synthetic deposit.
type connector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	cfg         config
	credentials string
}

// Sync checks that the credential is mounted and returns the deposit in
// transactions mode, nothing in opening mode, with the watermark it got.
func (c *connector) Sync(_ context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	if err := c.credential(); err != nil {
		return nil, status(connect.CodeFailedPrecondition, err)
	}
	d := &connectorv1.Delivery{BatchId: req.Msg.GetBatchId(), Institution: institution}
	if req.Msg.GetMode() == connectorv1.Mode_MODE_TRANSACTIONS && c.cfg.Account != "" {
		d.Transactions = []*connectorv1.Transaction{deposit(c.cfg)}
	}
	return connect.NewResponse(&connectorv1.SyncResponse{Delivery: d, Watermark: req.Msg.GetWatermark()}), nil
}

// status is the Connect error with code for err.
func status(code connect.Code, err error) *connect.Error { return connect.NewError(code, err) }

// credential checks that the mounted Secret holds a credential; its value
// never leaves this function.
func (c *connector) credential() error {
	data, err := os.ReadFile(filepath.Join(c.credentials, mountFile))
	if err != nil {
		return fmt.Errorf("%w: %w", errNoCredential, err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return errNoCredential
	}
	return nil
}

// deposit is the synthetic deposit into cfg's account on cfg's date: cash
// in, against equity (D-33).
func deposit(cfg config) *connectorv1.Transaction {
	amount := func(sign string) *connectorv1.Money {
		return &connectorv1.Money{Amount: sign + depositAmount, Currency: depositCurrency}
	}
	return &connectorv1.Transaction{
		Key:       &connectorv1.TransactionKey{Source: depositSource, ExternalId: depositPrefix + cfg.Date},
		AccountId: cfg.Account,
		Kind:      connectorv1.TransactionKind_TRANSACTION_KIND_DEPOSIT,
		Date:      cfg.Date,
		Postings: []*connectorv1.Posting{
			{Book: &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_CASH, Currency: depositCurrency}, Amount: amount("")},
			{Book: &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_EQUITY, Currency: depositCurrency}, Amount: amount("-")},
		},
	}
}
