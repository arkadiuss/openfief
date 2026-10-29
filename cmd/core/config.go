package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // the distroless image has no zoneinfo

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// inClusterEnv is set in every pod; its presence turns connectors on by
// default.
const inClusterEnv = "KUBERNETES_SERVICE_HOST"

// dsnEnv names the environment variable with the database's connection
// string. In the cluster it comes from CloudNativePG's app secret (D-29).
const dsnEnv = "OPENFIEF_DSN"

// The commands.
const (
	cmdServe       = "serve"
	cmdMigrate     = "migrate"
	cmdTokenIssue  = "token issue"
	cmdTokenList   = "token list"
	cmdTokenRevoke = "token revoke"
	cmdBind        = "connector bind"
	cmdSync        = "connector sync"
	cmdEnrollLink  = "enroll-link"
)

// Defaults of the serve flags.
const (
	defaultListen          = ":8080"
	defaultTimeZone        = "Europe/Warsaw"
	defaultEndOfDay        = "23:30"
	defaultWorkers         = 2
	defaultShutdownTimeout = 25 * time.Second
	defaultTokenTTL        = 720 * time.Hour
	defaultSyncEvery       = 6 * time.Hour
	defaultSyncJitter      = 30 * time.Minute
	defaultSyncTimeout     = 5 * time.Minute
	defaultRetryFirst      = 5 * time.Minute
	defaultRetryMax        = time.Hour
	defaultServiceAccount  = "openfief-core"
	defaultAuthOrigin      = "http://localhost:8080"
	defaultEnrollHost      = "localhost:8080"
)

var (
	// errUsage reports a command line the core does not understand.
	errUsage = errors.New("usage: core [serve|migrate|token issue|token list|token revoke ID|connector bind|connector sync|enroll-link] [flags]")
	// errNoDSN reports a missing connection string.
	errNoDSN = errors.New(dsnEnv + " is not set")
	// errNoInstance reports a connector command without --instance.
	errNoInstance = errors.New("--instance is required")
	// errNoAccount reports connector bind without --account.
	errNoAccount = errors.New("--account is required")
	// errBadMode reports a --mode that is not an ingest mode.
	errBadMode = errors.New("--mode must be opening or transactions")
	// errBadEndOfDay reports an end-of-day time that is not HH:MM.
	errBadEndOfDay = errors.New("end of day must be HH:MM, 00:00 to 23:59")
)

// command is one parsed command line.
type command struct {
	// Name is one of the cmd* constants.
	Name string
	DSN  string
	// Serve is what serve needs.
	Serve serveConfig
	// TokenName and TTL are token issue's; TokenID is token revoke's.
	TokenName string
	TTL       time.Duration
	TokenID   string
	// Connector is what connector bind and connector sync need.
	Connector connectorArgs
	// Enroll is what enroll-link needs.
	Enroll enrollArgs
}

// enrollArgs are enroll-link's flags: the origin the passkey is for, and
// how long the link lasts.
type enrollArgs struct {
	Origin identity.Origin
	TTL    time.Duration
}

// connectorArgs are the connector commands' flags.
type connectorArgs struct {
	Instance string
	// Account is connector bind's account.
	Account ledger.AccountID
	// Mode is connector sync's ingest mode; zero lets the instance's state
	// decide.
	Mode ingest.Mode
}

// serveConfig is what D-36 and D-62 make configurable, and where to listen.
type serveConfig struct {
	Listen string
	// Reporting is the reporting currency (D-62).
	Reporting kernel.Currency
	// EndOfDay is when the end-of-day revaluation runs; its Location is also
	// the time zone of "today".
	EndOfDay valuationpg.DailyAt
	Workers  int
	// PriceCheckThreshold is the relative difference beyond which a price
	// check is flagged (D-22).
	PriceCheckThreshold kernel.Decimal
	// EnsureDefaults stores the built-in sleeve kinds and sleeves at start.
	EnsureDefaults  bool
	ShutdownTimeout time.Duration
	// Connectors runs the connector scheduler (D-49); it needs the cluster.
	Connectors bool
	Sync       syncConfig
	// Auth is the AuthService's configuration.
	Auth authConfig
}

// syncConfig is the connector scheduler's configuration.
type syncConfig struct {
	Policy ingest.SyncPolicy
	// Timeout bounds one sync attempt.
	Timeout time.Duration
	// ServiceAccount is the core's own ServiceAccount, whose tokens it
	// presents to plugins (D-59).
	ServiceAccount string
}

// parseArgs reads the command line args and the environment through
// getenv. With no command, or only flags, it is serve.
func parseArgs(args []string, getenv func(string) string) (command, error) {
	cl := splitCommand(args)
	name, rest := cl.name, cl.args
	cmd := command{Name: name, DSN: getenv(dsnEnv)}
	var err error
	switch name {
	case cmdServe:
		cmd.Serve, err = parseServe(rest, getenv)
	case cmdMigrate, cmdTokenList:
		err = parseNone(name, rest)
	case cmdTokenIssue:
		cmd.TokenName, cmd.TTL, err = parseIssue(rest)
	case cmdTokenRevoke:
		cmd.TokenID, err = parseRevoke(rest)
	case cmdBind, cmdSync:
		cmd.Connector, err = parseConnector(name, rest)
	case cmdEnrollLink:
		cmd.Enroll, err = parseEnrollLink(rest)
	default:
		err = fmt.Errorf("%w: unknown command %q", errUsage, name)
	}
	if err != nil {
		return command{}, err
	}
	if cmd.DSN == "" {
		return command{}, errNoDSN
	}
	return cmd, nil
}

// commandLine is a command's name and its own arguments.
type commandLine struct {
	name string
	args []string
}

// splitCommand splits args into the command's name and its own arguments.
func splitCommand(args []string) commandLine {
	switch {
	case len(args) == 0 || strings.HasPrefix(args[0], "-"):
		return commandLine{name: cmdServe, args: args}
	case (args[0] == "token" || args[0] == "connector") && len(args) > 1:
		return commandLine{name: args[0] + " " + args[1], args: args[2:]}
	}
	return commandLine{name: args[0], args: args[1:]}
}

// newFlags returns an empty flag set for the command name that reports
// errors only through its return values.
func newFlags(name string) *flag.FlagSet {
	fs := flag.NewFlagSet("core "+name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

// parseServe parses serve's flags, reading the environment through getenv:
// in a cluster --connectors is on by default.
func parseServe(args []string, getenv func(string) string) (serveConfig, error) {
	inCluster := getenv(inClusterEnv) != ""
	fs := newFlags(cmdServe)
	c := serveConfig{}
	fs.StringVar(&c.Listen, "listen", defaultListen, "address of the read API and the health endpoints")
	reporting := fs.String("reporting-currency", jurisdictionPack().ReportingCurrency().String(), "the reporting currency (D-62); the jurisdiction pack's by default")
	zone := fs.String("time-zone", defaultTimeZone, "time zone of today and of the end of day")
	eod := fs.String("end-of-day", defaultEndOfDay, "when the end-of-day revaluation runs, HH:MM")
	fs.IntVar(&c.Workers, "workers", defaultWorkers, "revaluations that run at once (D-36)")
	threshold := fs.String("price-check-threshold", valuation.DefaultPriceCheckThreshold().String(),
		"relative difference beyond which a cross-check price is flagged, as a fraction: 0.01 is 1% (D-22)")
	fs.BoolVar(&c.EnsureDefaults, "ensure-defaults", true, "store the built-in sleeve kinds and sleeves that are missing")
	fs.DurationVar(&c.ShutdownTimeout, "shutdown-timeout", defaultShutdownTimeout, "how long requests and revaluations may drain on shutdown")
	fs.BoolVar(&c.Connectors, "connectors", inCluster, "sync the cluster's connector instances (D-49); on by default in a cluster")
	fs.DurationVar(&c.Sync.Policy.Interval, "sync-every", defaultSyncEvery, "how often each connector instance syncs")
	fs.DurationVar(&c.Sync.Policy.Jitter, "sync-jitter", defaultSyncJitter, "up to how much later than --sync-every a sync may start")
	fs.DurationVar(&c.Sync.Policy.RetryFirst, "sync-retry-first", defaultRetryFirst, "the first retry delay after a failed sync; it doubles per failure")
	fs.DurationVar(&c.Sync.Policy.RetryMax, "sync-retry-max", defaultRetryMax, "the longest retry delay")
	fs.DurationVar(&c.Sync.Timeout, "sync-timeout", defaultSyncTimeout, "how long one sync, call and import, may take")
	fs.StringVar(&c.Sync.ServiceAccount, "service-account", defaultServiceAccount, "the core's own ServiceAccount, whose tokens it presents to plugins (D-59)")
	origins := authFlags(fs, &c.Auth)
	if err := fs.Parse(args); err != nil {
		return serveConfig{}, fmt.Errorf("serve: %w", err)
	}
	if err := parseAuth(&c.Auth, *origins); err != nil {
		return serveConfig{}, err
	}
	if fs.NArg() > 0 {
		return serveConfig{}, fmt.Errorf("%w: unexpected %q", errUsage, fs.Arg(0))
	}
	if err := c.Sync.Policy.Validate(); err != nil {
		return serveConfig{}, fmt.Errorf("sync flags: %w", err)
	}
	var err error
	if c.Reporting, err = kernel.ParseCurrency(*reporting); err != nil {
		return serveConfig{}, fmt.Errorf("--reporting-currency: %w", err)
	}
	if flagErr := parseValuationFlags(&c, *threshold); flagErr != nil {
		return serveConfig{}, flagErr
	}
	loc, err := time.LoadLocation(*zone)
	if err != nil {
		return serveConfig{}, fmt.Errorf("--time-zone: %w", err)
	}
	if c.EndOfDay, err = parseEndOfDay(*eod, loc); err != nil {
		return serveConfig{}, fmt.Errorf("--end-of-day: %w", err)
	}
	return c, nil
}

// authFlags defines serve's authentication flags on fs, into a, and returns
// the allowed origins' flag.
func authFlags(fs *flag.FlagSet, a *authConfig) *stringsFlag {
	var origins stringsFlag
	fs.Var(&origins, "auth-origin", "an origin the owner's browser may use, such as https://dev.example.com; repeat for more (default "+defaultAuthOrigin+")")
	fs.DurationVar(&a.Sessions.Idle, "session-idle", identity.DefaultSessionIdle, "how long a session lasts unused")
	fs.DurationVar(&a.Sessions.Max, "session-max", identity.DefaultSessionMax, "how long a session lasts at most")
	fs.BoolVar(&a.DemoSessions, "demo-sessions", false, "issue read-only sessions without credentials on a database with the demo marker")
	return &origins
}

// parseAuth checks the session lifetimes and parses the allowed origins
// into a, defaulting to defaultAuthOrigin.
func parseAuth(a *authConfig, origins []string) error {
	if err := a.Sessions.Validate(); err != nil {
		return fmt.Errorf("--session-idle, --session-max: %w", err)
	}
	if len(origins) == 0 {
		origins = []string{defaultAuthOrigin}
	}
	var err error
	if a.Origins, err = identity.NewOrigins(origins); err != nil {
		return fmt.Errorf("--auth-origin: %w", err)
	}
	return nil
}

// parseEnrollLink parses enroll-link's flags.
func parseEnrollLink(args []string) (enrollArgs, error) {
	fs := newFlags(cmdEnrollLink)
	host := fs.String("host", defaultEnrollHost, "the host[:port] (or origin) the owner opens the link on: http for localhost, https otherwise")
	var e enrollArgs
	fs.DurationVar(&e.TTL, "ttl", identity.DefaultEnrollmentTTL, "how long the link works, at most 24h")
	if err := fs.Parse(args); err != nil {
		return enrollArgs{}, fmt.Errorf("enroll-link: %w", err)
	}
	if fs.NArg() > 0 {
		return enrollArgs{}, fmt.Errorf("%w: unexpected %q", errUsage, fs.Arg(0))
	}
	var err error
	e.Origin, err = originForHost(*host)
	return e, err
}

// parseValuationFlags parses the price check threshold into c.
func parseValuationFlags(c *serveConfig, threshold string) error {
	var err error
	if c.PriceCheckThreshold, err = parseThreshold(threshold); err != nil {
		return fmt.Errorf("--price-check-threshold: %w", err)
	}
	return nil
}

// parseThreshold parses a price check threshold: a fraction greater than 0
// and less than 1.
func parseThreshold(s string) (kernel.Decimal, error) {
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("parsing: %w", err)
	}
	if _, err := valuation.CheckThreshold(d); err != nil {
		return kernel.Decimal{}, fmt.Errorf("checking: %w", err)
	}
	return d, nil
}

// parseEndOfDay parses HH:MM in loc.
func parseEndOfDay(s string, loc *time.Location) (valuationpg.DailyAt, error) {
	hh, mm, ok := strings.Cut(s, ":")
	h, hErr := strconv.Atoi(hh)
	m, mErr := strconv.Atoi(mm)
	if !ok || len(hh) != 2 || len(mm) != 2 || hErr != nil || mErr != nil || h > 23 || m > 59 || h < 0 || m < 0 {
		return valuationpg.DailyAt{}, fmt.Errorf("%w: %q", errBadEndOfDay, s)
	}
	return valuationpg.DailyAt{Hour: h, Minute: m, Location: loc}, nil
}

// parseIssue parses token issue's flags.
func parseIssue(args []string) (string, time.Duration, error) {
	fs := newFlags(cmdTokenIssue)
	name := fs.String("name", "", "the token's name, such as the script that uses it")
	ttl := fs.Duration("ttl", defaultTokenTTL, "how long the token is valid, at most 8760h")
	if err := fs.Parse(args); err != nil {
		return "", 0, fmt.Errorf("token issue: %w", err)
	}
	if fs.NArg() > 0 {
		return "", 0, fmt.Errorf("%w: unexpected %q", errUsage, fs.Arg(0))
	}
	return *name, *ttl, nil
}

// parseRevoke parses token revoke's one argument, the token ID.
func parseRevoke(args []string) (string, error) {
	fs := newFlags(cmdTokenRevoke)
	if err := fs.Parse(args); err != nil {
		return "", fmt.Errorf("token revoke: %w", err)
	}
	if fs.NArg() != 1 {
		return "", fmt.Errorf("%w: token revoke takes one token ID", errUsage)
	}
	return fs.Arg(0), nil
}

// parseConnector parses connector bind's and connector sync's flags.
func parseConnector(name string, args []string) (connectorArgs, error) {
	fs := newFlags(name)
	var c connectorArgs
	fs.StringVar(&c.Instance, "instance", "", "the connector instance, a Plugin's name")
	account := fs.String("account", "", "the ledger account ID to bind the instance to")
	mode := fs.String("mode", "", "opening or transactions; by default the instance's state decides")
	if err := fs.Parse(args); err != nil {
		return connectorArgs{}, fmt.Errorf("%s: %w", name, err)
	}
	switch {
	case fs.NArg() > 0:
		return connectorArgs{}, fmt.Errorf("%w: unexpected %q", errUsage, fs.Arg(0))
	case c.Instance == "":
		return connectorArgs{}, errNoInstance
	case name == cmdBind && *account == "":
		return connectorArgs{}, errNoAccount
	case name == cmdBind && *mode != "", name == cmdSync && *account != "":
		return connectorArgs{}, fmt.Errorf("%w: --account is bind's, --mode is sync's", errUsage)
	}
	c.Account = ledger.AccountID(*account)
	var err error
	c.Mode, err = parseMode(*mode)
	return c, err
}

// parseMode parses --mode: empty, opening or transactions.
func parseMode(s string) (ingest.Mode, error) {
	for _, m := range []ingest.Mode{ingest.ModeOpening, ingest.ModeTransactions} {
		if s == m.String() {
			return m, nil
		}
	}
	if s != "" {
		return 0, fmt.Errorf("%w: %q", errBadMode, s)
	}
	return 0, nil
}

// parseNone checks that the command name has no arguments.
func parseNone(name string, args []string) error {
	fs := newFlags(name)
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("%w: unexpected %q", errUsage, fs.Arg(0))
	}
	return nil
}
