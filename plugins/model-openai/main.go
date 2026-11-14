// Command model-openai is OpenFief's first-party model-provider plugin for
// the OpenAI-compatible chat protocol (D-205): the core's model router calls
// it to list the provider's models and to generate with them, with tools and
// structured output when a model's entry allows them. The protocol serves
// OpenAI, OpenRouter, Ollama, vLLM, LM Studio and others, so one image is
// installed once per provider instance, each with its own base URL, key,
// grant and egress allow-list entry (D-48). It reaches the provider only
// through the egress proxy and only the host its grant approves (D-32), and
// it never sees a role, a data tier or who asked: the core decides what may
// be sent (D-206) and the plugin sends it.
//
// The plugin keeps and logs no request or answer text (GR-12, D-211): an
// error names a status, never the provider's reply.
//
// The API key, when the provider needs one, comes from the instance's
// optional Secret <instance>--api-key (file api-key, D-89), mounted at
// /etc/openfief/model-provider. It is read on every call and never logged;
// a local provider such as Ollama has none. Config (OPENFIEF_CONFIG,
// checked against the manifest's schema):
//
//	{"baseUrl": "https://openrouter.ai/api/v1", "timeout": "60s",
//	 "location": "third_party_cloud", "retention": {"days": 30, "mayTrain": false},
//	 "models": [{"id": "openai/gpt-oss-120b", "inputPrice": "0.15", "outputPrice": "0.60",
//	             "priceCurrency": "USD", "contextTokens": 131072, "maxOutputTokens": 8192,
//	             "toolCalling": true, "structuredOutput": true}]}
//
// baseUrl is the provider's API root, such as https://api.openai.com/v1: an
// https URL whose host the install shows and the grant approves (D-75). A
// provider in the cluster, such as Ollama, is reached through an approved
// in-cluster service instead (D-74). location and retention say where the
// models run and how long the provider keeps what it receives: the owner's
// own statement, which the router shows and the owner's choices of which
// models may see which data rest on (D-206). Prices are per million tokens.
// The client trusts the system's roots plus the PEM certificates of the
// instance's optional Secret <instance>--ca-bundle (file ca.crt), mounted at
// /etc/openfief/ca-bundle.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/arkadiuss/openfief/plugins/model-openai/openaichat"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

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
		log.Error("model-openai stopped", "error", openfief.LogSafe(err,
			errBadConfig, errBadEntry, errNoKey, openfief.ErrBaseURL, openfief.ErrCABundle, openfief.ErrNoEgressProxy))
		return 1
	}
	return 0
}

// run serves the provider until ctx is done.
func run(ctx context.Context, log *slog.Logger) error {
	var raw rawConfig
	if err := openfief.LoadConfig(&raw); err != nil {
		return fmt.Errorf("config: %w", err)
	}
	cfg, err := raw.parse()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	ep, err := openfief.LoadEndpoint(raw.BaseURL, "")
	if err != nil {
		return fmt.Errorf("config baseUrl: %w", err)
	}
	hc, err := ep.NewClient()
	if err != nil {
		return fmt.Errorf("egress: %w", err)
	}
	// The per-call timeout is the context's, so a slow model is cut off
	// where the config says; the client's own bounds only the headers.
	log.Info("model-openai serving", "models", len(cfg.models), "extra_ca", len(ep.CABundle) > 0)
	svc := &provider{cfg: cfg, chat: &openaichat.Client{HTTP: hc, BaseURL: ep.BaseURL, Key: keyFromMount(keyMount)}}
	if err := openfief.ServeModelProvider(ctx, svc); err != nil {
		return fmt.Errorf("model-openai: %w", err)
	}
	return nil
}
