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

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

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
)

// Defaults of the serve flags.
const (
	defaultListen          = ":8080"
	defaultReporting       = "PLN"
	defaultTimeZone        = "Europe/Warsaw"
	defaultEndOfDay        = "23:30"
	defaultWorkers         = 2
	defaultShutdownTimeout = 25 * time.Second
	defaultTokenTTL        = 720 * time.Hour
)

var (
	// errUsage reports a command line the core does not understand.
	errUsage = errors.New("usage: core [serve|migrate|token issue|token list|token revoke ID] [flags]")
	// errNoDSN reports a missing connection string.
	errNoDSN = errors.New(dsnEnv + " is not set")
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
}

// serveConfig is what D-36 and D-62 make configurable, and where to listen.
type serveConfig struct {
	Listen string
	// Reporting is the reporting currency (D-62).
	Reporting kernel.Currency
	// EndOfDay is when the end-of-day revaluation runs; its Location is also
	// the time zone of "today".
	EndOfDay  valuationpg.DailyAt
	Workers   int
	BondBasis valuation.Basis
	// EnsureDefaults stores the built-in sleeve kinds and sleeves at start.
	EnsureDefaults  bool
	ShutdownTimeout time.Duration
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
		cmd.Serve, err = parseServe(rest)
	case cmdMigrate:
		err = parseNone(name, rest)
	case cmdTokenIssue:
		cmd.TokenName, cmd.TTL, err = parseIssue(rest)
	case cmdTokenList:
		err = parseNone(name, rest)
	case cmdTokenRevoke:
		cmd.TokenID, err = parseRevoke(rest)
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
	case args[0] == "token" && len(args) > 1:
		return commandLine{name: "token " + args[1], args: args[2:]}
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

// parseServe parses serve's flags.
func parseServe(args []string) (serveConfig, error) {
	fs := newFlags(cmdServe)
	c := serveConfig{}
	fs.StringVar(&c.Listen, "listen", defaultListen, "address of the read API and the health endpoints")
	reporting := fs.String("reporting-currency", defaultReporting, "the reporting currency (D-62)")
	zone := fs.String("time-zone", defaultTimeZone, "time zone of today and of the end of day")
	eod := fs.String("end-of-day", defaultEndOfDay, "when the end-of-day revaluation runs, HH:MM")
	fs.IntVar(&c.Workers, "workers", defaultWorkers, "revaluations that run at once (D-36)")
	basis := fs.String("bond-basis", valuation.DefaultBondBasis.String(), "bond basis for totals: gross_accrued or net_if_redeemed (D-25)")
	fs.BoolVar(&c.EnsureDefaults, "ensure-defaults", true, "store the built-in sleeve kinds and sleeves that are missing")
	fs.DurationVar(&c.ShutdownTimeout, "shutdown-timeout", defaultShutdownTimeout, "how long requests and revaluations may drain on shutdown")
	if err := fs.Parse(args); err != nil {
		return serveConfig{}, fmt.Errorf("serve: %w", err)
	}
	if fs.NArg() > 0 {
		return serveConfig{}, fmt.Errorf("%w: unexpected %q", errUsage, fs.Arg(0))
	}
	var err error
	if c.Reporting, err = kernel.ParseCurrency(*reporting); err != nil {
		return serveConfig{}, fmt.Errorf("--reporting-currency: %w", err)
	}
	if c.BondBasis, err = parseBondBasis(*basis); err != nil {
		return serveConfig{}, fmt.Errorf("--bond-basis: %w", err)
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

// parseBondBasis parses a basis name and checks that it can value bonds.
func parseBondBasis(s string) (valuation.Basis, error) {
	b, err := valuation.ParseBasis(s)
	if err != nil {
		return 0, fmt.Errorf("parsing: %w", err)
	}
	if _, err := valuation.BondBasis(b); err != nil {
		return 0, fmt.Errorf("checking: %w", err)
	}
	return b, nil
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
