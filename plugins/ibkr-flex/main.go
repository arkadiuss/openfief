// Command ibkr-flex is OpenFief's first-party Interactive Brokers Flex
// Query connector plugin (D-49): the core calls its Sync to pull the
// statement of the owner's Flex Query from IBKR's Flex Web Service, and its
// ImportFile to map an uploaded Flex statement, in the opening or
// transactions mode the call names (D-23). It reaches
// ndcdyn.interactivebrokers.com, or the base URL the owner configured, only
// through the egress proxy (D-32). The Flex parser and mapping are
// internal/ingest/ibkrflex; this binary only wires them to the SDK.
//
// The Flex Web Service token and the Flex Query's ID come from the instance's
// Secret <instance>--flex-token (role flex-token, D-89; files token and query-id), which the owner
// creates (D-29) and the manifest mounts at /etc/openfief/ibkr-flex. They
// are read on every sync and never logged.
//
// Config (OPENFIEF_CONFIG, checked against the manifest's schema) holds only
// baseUrl, optional: the Flex Web Service, or a mirror of it, an https URL
// whose host the install shows and the grant approves (D-75); SendRequest
// and GetStatement are below it. The client trusts the system's roots plus
// the PEM certificates of the instance's optional Secret
// <instance>--ca-bundle (file ca.crt), mounted at /etc/openfief/ca-bundle:
//
//	{"baseUrl": "https://ndcdyn.interactivebrokers.com/AccountManagement/FlexWebService"}
//
// The plugin names every statement's entries by the
// IBKR account the statement is of, and reports every IBKR account it saw,
// so the core maps them to the OpenFief accounts the owner picked in the
// built-in UI and books nothing for an account the owner hasn't mapped
// (D-72). The owner never types an IBKR account number, and the plugin
// never learns an OpenFief account ID.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest/ibkrflex"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// requestTimeout bounds one request to the Flex Web Service.
const requestTimeout = time.Minute

// defaultBaseURL is the Flex Web Service IBKR documents: its SendRequest
// and GetStatement endpoints are below it.
const defaultBaseURL = "https://ndcdyn.interactivebrokers.com/AccountManagement/FlexWebService"

// config is the plugin's configuration.
type config struct {
	BaseURL string `json:"baseUrl"`
}

func main() {
	os.Exit(serve())
}

// serve runs the plugin until SIGTERM or an interrupt and returns its exit
// code.
func serve() int {
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()
	if err := run(ctx, log); err != nil {
		log.Error("ibkr-flex stopped", "error", err.Error())
		return 1
	}
	return 0
}

// run serves the connector until ctx is done.
func run(ctx context.Context, log *slog.Logger) error {
	var cfg config
	if err := openfief.LoadConfig(&cfg); err != nil {
		return fmt.Errorf("config: %w", err)
	}
	ep, err := openfief.LoadEndpoint(cfg.BaseURL, defaultBaseURL)
	if err != nil {
		return fmt.Errorf("config baseUrl: %w", err)
	}
	hc, err := ep.NewClient()
	if err != nil {
		return fmt.Errorf("egress: %w", err)
	}
	hc.Timeout = requestTimeout
	c := &connector{
		mapper:  ibkrflex.Mapper{ByBrokerAccount: true},
		secrets: flexMount,
		fetch:   flexFetch(hc, ep.BaseURL),
	}
	log.Info("ibkr-flex serving", "extra_ca", len(ep.CABundle) > 0)
	if err := openfief.ServeConnector(ctx, c); err != nil {
		return fmt.Errorf("ibkr-flex: %w", err)
	}
	return nil
}

// flexFetch fetches the statement of a Flex Query from the Flex Web Service
// at baseURL with hc.
func flexFetch(hc *http.Client, baseURL string) func(ctx context.Context, token, queryID string) ([]byte, error) {
	urls := flexEndpoints(baseURL)
	return func(ctx context.Context, token, queryID string) ([]byte, error) {
		client := ibkrflex.Client{
			HTTP: hc, Token: token, QueryID: queryID, SendRequestURL: urls.sendRequest, GetStatementURL: urls.getStatement,
		}
		return client.FetchStatement(ctx)
	}
}

// flexURLs are the two endpoints of the Flex Web Service.
type flexURLs struct {
	sendRequest  string
	getStatement string
}

// flexEndpoints are the SendRequest and GetStatement endpoints below the
// Flex Web Service's baseURL.
func flexEndpoints(baseURL string) flexURLs {
	return flexURLs{sendRequest: baseURL + "/SendRequest", getStatement: baseURL + "/GetStatement"}
}
