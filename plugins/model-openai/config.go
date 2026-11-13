package main

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"time"

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
)

// A decimal price: digits, optionally a fraction. A function-free variable
// of the package main, not a domain package.
var (
	decimalPattern  = regexp.MustCompile(`^\d{1,12}(\.\d{1,12})?$`)
	currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)
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
}

// config is the checked configuration of one provider instance (D-205).
type config struct {
	timeout   time.Duration
	location  modelsv1.Location
	retention *modelsv1.Retention
	models    []*modelsv1.ModelEntry
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

// entry returns the entry with id.
func (c *config) entry(id string) (*modelsv1.ModelEntry, bool) {
	for _, e := range c.models {
		if e.GetId() == id {
			return e, true
		}
	}
	return nil, false
}
