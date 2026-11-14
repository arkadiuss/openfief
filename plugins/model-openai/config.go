package main

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/plugins/model-openai/openaichat"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
)

// Errors reading the config. Each wraps the reason and is for the plugin's
// log and the owner: it names fields, never a Secret.
// errBadConfig reports a config the manifest's schema would refuse, or one
// the schema can't check, such as a model listed twice.
var errBadConfig = errors.New("invalid config")

const (
	// defaultTimeout bounds one call to the provider when the config sets
	// none: slow models may think for a minute.
	defaultTimeout = time.Minute
	// maxTimeout is the longest timeout the config may set, below the
	// core's own bound on a call to a plugin.
	maxTimeout = 2 * time.Minute
	// maxModels is the most model entries a provider instance holds.
	maxModels = 64
	// maxHosts is the most hosts an entry pins, and maxHostLen the longest
	// slug of one.
	maxHosts   = 8
	maxHostLen = 64
)

// A decimal price: digits, optionally a fraction. A function-free variable
// of the package main, not a domain package.
var (
	decimalPattern  = regexp.MustCompile(`^\d{1,12}(\.\d{1,12})?$`)
	currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)
	// hostPattern is an OpenRouter provider slug, such as "groq" or
	// "nextbit/bf16": lower-case letters, digits and . _ - /.
	hostPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._/-]*$`)
)

// rawConfig is the plugin's configuration as OPENFIEF_CONFIG holds it.
type rawConfig struct {
	BaseURL   string       `json:"baseUrl"`
	Timeout   string       `json:"timeout"`
	Location  string       `json:"location"`
	Retention rawRetention `json:"retention"`
	Models    []rawModel   `json:"models"`
}

type rawRetention struct {
	Days     int32 `json:"days"`
	MayTrain bool  `json:"mayTrain"`
}

type rawModel struct {
	ID               string `json:"id"`
	InputPrice       string `json:"inputPrice"`
	OutputPrice      string `json:"outputPrice"`
	PriceCurrency    string `json:"priceCurrency"`
	ContextTokens    int64  `json:"contextTokens"`
	MaxOutputTokens  int64  `json:"maxOutputTokens"`
	ToolCalling      bool   `json:"toolCalling"`
	StructuredOutput bool   `json:"structuredOutput"`
	// Hosts pins the entry to OpenRouter's providers of these slugs, in this
	// order and with no fallback to others; Reasoning sets OpenRouter's
	// reasoning control (off, low, medium or high); DataCollection sets
	// OpenRouter's provider.data_collection (allow or deny: deny lets no host
	// that may store the request or train on it serve it). All three are
	// OpenRouter's, and a config that sets one for another server is refused.
	Hosts          []string `json:"hosts"`
	Reasoning      string   `json:"reasoning"`
	DataCollection string   `json:"dataCollection"`
}

// config is the checked configuration of one provider instance (D-205).
type config struct {
	timeout   time.Duration
	location  modelsv1.Location
	retention *modelsv1.Retention
	models    []*modelsv1.ModelEntry
	// routing holds the OpenRouter options of the entries that set any, by
	// the entry's id.
	routing map[string]routing
}

// routing is one entry's OpenRouter options: the hosts it is pinned to, its
// reasoning setting and its data collection setting, each empty when the entry
// sets none.
type routing struct {
	hosts          []string
	reasoning      string
	dataCollection string
}

// parse checks the raw config: a known location, a timeout in range, one
// to maxModels entries with distinct ids, and a price in a currency or none.
func (r *rawConfig) parse() (config, error) {
	timeout, err := r.timeout()
	if err != nil {
		return config{}, err
	}
	c := config{timeout: timeout}
	switch r.Location {
	case "owner_hardware":
		c.location = modelsv1.Location_LOCATION_OWNER_HARDWARE
	case "third_party_cloud":
		c.location = modelsv1.Location_LOCATION_THIRD_PARTY_CLOUD
	default:
		return config{}, fmt.Errorf("%w: location is owner_hardware or third_party_cloud", errBadConfig)
	}
	if r.Retention.Days < 0 {
		return config{}, fmt.Errorf("%w: retention days is not negative", errBadConfig)
	}
	c.retention = &modelsv1.Retention{Days: r.Retention.Days, MayTrain: r.Retention.MayTrain}
	c.models, err = r.entries()
	if err != nil {
		return config{}, err
	}
	c.routing, err = r.routings()
	if err != nil {
		return config{}, err
	}
	return c, nil
}

// timeout is the configured timeout, or defaultTimeout; it must be a
// duration up to maxTimeout.
func (r *rawConfig) timeout() (time.Duration, error) {
	if r.Timeout == "" {
		return defaultTimeout, nil
	}
	d, err := time.ParseDuration(r.Timeout)
	if err != nil || d <= 0 || d > maxTimeout {
		return 0, fmt.Errorf("%w: timeout is a duration up to %s", errBadConfig, maxTimeout)
	}
	return d, nil
}

// entries checks the models: one to maxModels, each valid, no id twice.
func (r *rawConfig) entries() ([]*modelsv1.ModelEntry, error) {
	if len(r.Models) == 0 || len(r.Models) > maxModels {
		return nil, fmt.Errorf("%w: one to %d models", errBadConfig, maxModels)
	}
	var out []*modelsv1.ModelEntry
	var ids []string
	for i := range r.Models {
		e, err := r.Models[i].entry()
		if err != nil {
			return nil, fmt.Errorf("%w: models[%d]: %w", errBadConfig, i, err)
		}
		if slices.Contains(ids, e.GetId()) {
			return nil, fmt.Errorf("%w: models[%d]: id listed twice", errBadConfig, i)
		}
		ids = append(ids, e.GetId())
		out = append(out, e)
	}
	return out, nil
}

// routings checks the OpenRouter options of the entries: pinned hosts, a
// reasoning setting and a data collection setting are OpenRouter's, so they
// are refused on any other base URL, where the plugin would never send them.
func (r *rawConfig) routings() (map[string]routing, error) {
	onOpenRouter := (&openaichat.Client{BaseURL: r.BaseURL}).OnOpenRouter()
	var out map[string]routing
	for i := range r.Models {
		m := &r.Models[i]
		if len(m.Hosts) == 0 && m.Reasoning == "" && m.DataCollection == "" {
			continue
		}
		if !onOpenRouter {
			return nil, fmt.Errorf("%w: models[%d]: hosts, reasoning and dataCollection are OpenRouter's: the baseUrl's host is openrouter.ai", errBadConfig, i)
		}
		if err := m.checkRouting(); err != nil {
			return nil, fmt.Errorf("%w: models[%d]: %w", errBadConfig, i, err)
		}
		if out == nil {
			out = make(map[string]routing)
		}
		out[m.ID] = routing{hosts: slices.Clone(m.Hosts), reasoning: m.Reasoning, dataCollection: m.DataCollection}
	}
	return out, nil
}

// checkRouting checks the entry's hosts, at most maxHosts distinct provider
// slugs, and its reasoning setting.
func (m *rawModel) checkRouting() error {
	if len(m.Hosts) > maxHosts {
		return fmt.Errorf("%w: at most %d hosts", errBadEntry, maxHosts)
	}
	for i, h := range m.Hosts {
		if len(h) > maxHostLen || !hostPattern.MatchString(h) {
			return fmt.Errorf("%w: a host is an OpenRouter provider slug", errBadEntry)
		}
		if slices.Contains(m.Hosts[:i], h) {
			return fmt.Errorf("%w: a host listed twice", errBadEntry)
		}
	}
	if m.Reasoning != "" && !openaichat.ValidReasoning(m.Reasoning) {
		return fmt.Errorf("%w: reasoning is off, low, medium or high", errBadEntry)
	}
	if m.DataCollection != "" && !openaichat.ValidDataCollection(m.DataCollection) {
		return fmt.Errorf("%w: dataCollection is allow or deny", errBadEntry)
	}
	return nil
}

// errBadEntry reports a model entry the schema would refuse.
var errBadEntry = errors.New("invalid model entry")

// entry checks one model and returns its entry: an id, both prices or
// neither, and non-negative limits.
func (m *rawModel) entry() (*modelsv1.ModelEntry, error) {
	if m.ID == "" || len(m.ID) > 128 {
		return nil, fmt.Errorf("%w: an id of one to 128 characters", errBadEntry)
	}
	if m.pricedWrongly() {
		return nil, fmt.Errorf("%w: a price is both prices, decimal strings, and an ISO 4217 currency", errBadEntry)
	}
	if m.ContextTokens < 0 || m.MaxOutputTokens < 0 {
		return nil, fmt.Errorf("%w: limits are not negative", errBadEntry)
	}
	return &modelsv1.ModelEntry{
		Id: m.ID, InputPrice: m.InputPrice, OutputPrice: m.OutputPrice, PriceCurrency: m.PriceCurrency,
		ContextTokens: m.ContextTokens, MaxOutputTokens: m.MaxOutputTokens,
		ToolCalling: m.ToolCalling, StructuredOutput: m.StructuredOutput,
	}, nil
}

// pricedWrongly reports a price that is part of one: some of the two prices
// and the currency, but not both prices as decimals and a currency.
func (m *rawModel) pricedWrongly() bool {
	if m.InputPrice == "" && m.OutputPrice == "" && m.PriceCurrency == "" {
		return false
	}
	return !decimalPattern.MatchString(m.InputPrice) || !decimalPattern.MatchString(m.OutputPrice) ||
		!currencyPattern.MatchString(m.PriceCurrency)
}

// chatFor returns the chat client that serves the entry with id: base, with
// the entry's pinned hosts, reasoning and data collection settings when it has them. base is
// not changed. The client sends them to OpenRouter only.
func (c *config) chatFor(base *openaichat.Client, id string) *openaichat.Client {
	r, ok := c.routing[id]
	if !ok {
		return base
	}
	out := *base
	out.Route, out.Reasoning, out.DataCollection = r.hosts, r.reasoning, r.dataCollection
	return &out
}

// entry returns the entry with id.
func (c *config) entry(id string) (*modelsv1.ModelEntry, bool) {
	for _, e := range c.models {
		if e.GetId() == id {
			return e, true
		}
	}
	return nil, false
}
