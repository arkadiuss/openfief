package main

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/modelrouter"
	modelsapp "github.com/arkadiuss/openfief/internal/modelrouter/app"
	modelspg "github.com/arkadiuss/openfief/internal/modelrouter/postgres"
	"github.com/arkadiuss/openfief/internal/views/browserproxy"
)

// errGuards reports a guard setting the core can't use.
var errGuards = errors.New("invalid model guard setting")

// guardConfig is D-208's guards on the model router, set by flags. Every one is
// off in its zero value, which is what an installation that sets none gets: the
// demo sets them (deploy/demo/model).
type guardConfig struct {
	// DailyBudgetUSD is the most the day's model calls may cost, in US dollars,
	// as decimal text; empty or zero is no budget.
	DailyBudgetUSD string
	// QuestionsPerHour and QuestionsPerDay limit the Ask box's questions to a
	// model per session and per address; zero is no limit.
	QuestionsPerHour int
	QuestionsPerDay  int
	// AskCalls is how many calls of the Ask box's role may be in flight at once
	// in this process, and QueueWait how long a call past it waits for a free
	// one before it is refused as busy.
	AskCalls  int
	QueueWait time.Duration
	// ClientIPHeader and TrustedProxies say whose header the front door believes
	// for the browser's address; AllowTurnstile widens the content security
	// policy of the pages it proxies to Cloudflare Turnstile's script and frame.
	ClientIPHeader string
	TrustedProxies string
	AllowTurnstile bool
}

// guardFlags defines serve's guard flags on fs, into g.
func guardFlags(fs *flag.FlagSet, g *guardConfig) {
	fs.StringVar(&g.DailyBudgetUSD, "model-daily-budget-usd", "",
		"the most the day's model calls may cost, in US dollars, counted from the model entries' prices (D-208); empty is no budget. The day is --time-zone's")
	fs.IntVar(&g.QuestionsPerHour, "ask-questions-per-hour", 0, "the most questions to a model one Ask box session, and one address, may ask in a rolling hour (D-208); 0 is no limit")
	fs.IntVar(&g.QuestionsPerDay, "ask-questions-per-day", 0, "the most questions to a model one Ask box session, and one address, may ask in a rolling day (D-208); 0 is no limit")
	fs.IntVar(&g.AskCalls, "ask-concurrent-calls", 0, "the most calls of the Ask box's role in flight at once in this process (D-208); 0 is no limit")
	fs.DurationVar(&g.QueueWait, "ask-queue-wait", modelsapp.DefaultQueueWait, "how long an Ask box call waits for a free one of --ask-concurrent-calls before it is refused as busy")
	fs.StringVar(&g.ClientIPHeader, "client-ip-header", "",
		"the header of the reverse proxy in front of the core that holds the browser's address, X-Forwarded-For or one with a single address such as CF-Connecting-IP; believed only from --trusted-proxies")
	fs.StringVar(&g.TrustedProxies, "trusted-proxies", "", "the addresses and networks, CIDR, comma separated, of the proxies whose --client-ip-header is believed, such as the Ingress controller's pods")
	fs.BoolVar(&g.AllowTurnstile, "allow-turnstile", false,
		"allow Cloudflare Turnstile's script and frame, and nothing else from outside, in the pages of view plugins (the demo's Ask box widget, D-208)")
}

// validate checks the guard settings together: a budget that parses and isn't
// negative, limits that are, an hourly limit not above the daily one, and a
// client address header with the proxies it is trusted from.
func (g *guardConfig) validate() error {
	if _, err := g.budget(); err != nil {
		return err
	}
	if err := g.limits().Validate(); err != nil {
		return fmt.Errorf("%w: --ask-questions-per-hour and --ask-questions-per-day: %w", errGuards, err)
	}
	if g.AskCalls < 0 {
		return fmt.Errorf("%w: --ask-concurrent-calls is not negative", errGuards)
	}
	if g.QueueWait < 0 {
		return fmt.Errorf("%w: --ask-queue-wait is not negative", errGuards)
	}
	if _, err := browserproxy.ParseClientIPConfig(g.ClientIPHeader, g.TrustedProxies); err != nil {
		return fmt.Errorf("--client-ip-header, --trusted-proxies: %w", err)
	}
	return nil
}

// budget returns the daily budget in US dollars: none when the flag is empty.
func (g *guardConfig) budget() (modelrouter.Budget, error) {
	if g.DailyBudgetUSD == "" {
		return modelrouter.Budget{}, nil
	}
	limit, err := kernel.ParseDecimal(g.DailyBudgetUSD)
	if err != nil {
		return modelrouter.Budget{}, fmt.Errorf("%w: --model-daily-budget-usd is a decimal such as 0.50: %w", errGuards, err)
	}
	usd, err := kernel.ParseCurrency("USD")
	if err != nil {
		return modelrouter.Budget{}, fmt.Errorf("the budget's currency: %w", err)
	}
	b, err := modelrouter.NewBudget(limit, usd)
	if err != nil {
		return modelrouter.Budget{}, fmt.Errorf("%w: --model-daily-budget-usd: %w", errGuards, err)
	}
	return b, nil
}

// limits returns the limits of questions per session and per address.
func (g *guardConfig) limits() modelrouter.QuestionLimits {
	return modelrouter.QuestionLimits{PerHour: g.QuestionsPerHour, PerDay: g.QuestionsPerDay}
}

// routerGuards returns the router's guards over the usage store, counting the day
// in zone, which is the core's own time zone.
func (g *guardConfig) routerGuards(usage *modelspg.Store, zone *time.Location) (modelsapp.Guards, error) {
	b, err := g.budget()
	if err != nil {
		return modelsapp.Guards{}, err
	}
	return modelsapp.Guards{Budget: b, Usage: usage, Zone: zone, AskCalls: g.AskCalls, QueueWait: g.QueueWait}, nil
}

// proxyOptions returns the front door's options: the browser's address as the
// guards read it, and the content security policy.
func (g *guardConfig) proxyOptions() ([]browserproxy.Option, error) {
	c, err := browserproxy.ParseClientIPConfig(g.ClientIPHeader, g.TrustedProxies)
	if err != nil {
		return nil, fmt.Errorf("--client-ip-header, --trusted-proxies: %w", err)
	}
	opts := []browserproxy.Option{browserproxy.WithClientIP(c)}
	if g.AllowTurnstile {
		opts = append(opts, browserproxy.WithTurnstile())
	}
	return opts, nil
}
