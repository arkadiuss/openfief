// Package connectrpc calls model-provider plugins' contract,
// openfief.modelrouter.v1.ModelProviderService, over Connect RPC (D-205): the
// core's side of a model provider. It authenticates each call as the core
// with a token whose audience is the instance (D-59), and translates the
// contract's messages into the domain, so they never reach it. It logs
// nothing: a request's text is the owner's data (GR-12).
package connectrpc

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
)

const (
	// MaxResponseBytes caps a plugin's answer: a model's answer and its
	// tool calls, far above what a role asks for.
	MaxResponseBytes = 4 << 20
	// CallTimeout bounds one call to a plugin, above the plugin's own
	// timeout for its provider (the longest a slow model may take).
	CallTimeout = 3 * time.Minute
)

// Tokens issues the token the core presents to a plugin instance: a
// short-lived ServiceAccount token whose audience is that instance (D-59).
// The identity context implements it with the TokenRequest API.
type Tokens interface {
	TokenFor(ctx context.Context, instance string) (string, error)
}

// Client is app.Providers over Connect RPC.
type Client struct {
	http    connect.HTTPClient
	baseURL func(instance string) (string, error)
	tokens  Tokens
}

var _ app.Providers = (*Client)(nil)

// NewClient returns a client that reaches a provider instance at the URL
// baseURL returns for it, such as its ClusterIP Service, with httpClient,
// and authenticates with a token from tokens.
func NewClient(httpClient connect.HTTPClient, baseURL func(instance string) (string, error), tokens Tokens) *Client {
	return &Client{http: httpClient, baseURL: baseURL, tokens: tokens}
}

// Describe returns where the provider's models run and its retention terms.
func (c *Client) Describe(ctx context.Context, provider string) (modelrouter.Provider, error) {
	t, err := c.service(ctx, provider)
	if err != nil {
		return modelrouter.Provider{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, CallTimeout)
	defer cancel()
	req := connect.NewRequest(&modelsv1.DescribeRequest{})
	req.Header().Set("Authorization", t.auth)
	resp, err := t.svc.Describe(ctx, req)
	if err != nil {
		return modelrouter.Provider{}, callError(provider, err)
	}
	return providerFrom(resp.Msg)
}

// Models lists the provider's model entries.
func (c *Client) Models(ctx context.Context, provider string) ([]modelrouter.Entry, error) {
	t, err := c.service(ctx, provider)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, CallTimeout)
	defer cancel()
	req := connect.NewRequest(&modelsv1.ListModelsRequest{})
	req.Header().Set("Authorization", t.auth)
	resp, err := t.svc.ListModels(ctx, req)
	if err != nil {
		return nil, callError(provider, err)
	}
	out := make([]modelrouter.Entry, 0, len(resp.Msg.GetModels()))
	for _, m := range resp.Msg.GetModels() {
		e, err := entryFrom(m)
		if err != nil {
			return nil, fmt.Errorf("provider %s: %w: %w", provider, app.ErrProviderRejected, err)
		}
		out = append(out, e)
	}
	return out, nil
}

// Generate asks the provider's model entry for an answer to req.
func (c *Client) Generate(ctx context.Context, provider, model string, req *modelrouter.Request) (modelrouter.Result, error) {
	t, err := c.service(ctx, provider)
	if err != nil {
		return modelrouter.Result{}, err
	}
	pb, err := requestTo(model, req)
	if err != nil {
		return modelrouter.Result{}, fmt.Errorf("%w: %w", app.ErrProviderRejected, err)
	}
	ctx, cancel := context.WithTimeout(ctx, CallTimeout)
	defer cancel()
	call := connect.NewRequest(pb)
	call.Header().Set("Authorization", t.auth)
	resp, err := t.svc.Generate(ctx, call)
	if err != nil {
		return modelrouter.Result{}, callError(provider, err)
	}
	return resultFrom(resp.Msg), nil
}

// service returns the generated client for provider and the Authorization
// header value with a fresh core token. An instance that has no address or
// no token is unavailable.
func (c *Client) service(ctx context.Context, provider string) (target, error) {
	url, err := c.baseURL(provider)
	if err != nil {
		return target{}, fmt.Errorf("%w: provider %s address: %w", app.ErrProviderUnavailable, provider, err)
	}
	token, err := c.tokens.TokenFor(ctx, provider)
	if err != nil {
		return target{}, fmt.Errorf("%w: core token for %s: %w", app.ErrProviderUnavailable, provider, err)
	}
	return target{
		svc:  modelsv1connect.NewModelProviderServiceClient(c.http, url, connect.WithReadMaxBytes(MaxResponseBytes)),
		auth: "Bearer " + token,
	}, nil
}

// target is a provider instance's generated client and the Authorization
// header value that authenticates the core to it.
type target struct {
	svc  modelsv1connect.ModelProviderServiceClient
	auth string
}

// callError classifies a failed call by its Connect code, never by its text
// (which may quote the provider's reply): a request the plugin rejected is
// app.ErrProviderRejected, a context's end stays as it is, and anything else,
// an unreachable, slow or failing provider, is app.ErrProviderUnavailable.
// The message names the provider and the code only.
func callError(provider string, err error) error {
	if errors.Is(err, context.Canceled) {
		return err
	}
	code := connect.CodeOf(err)
	rejecting := []connect.Code{
		connect.CodeInvalidArgument, connect.CodeNotFound, connect.CodeFailedPrecondition, connect.CodeUnimplemented, connect.CodeOutOfRange,
	}
	if slices.Contains(rejecting, code) {
		return fmt.Errorf("provider %s answered %s: %w", provider, code, app.ErrProviderRejected)
	}
	return fmt.Errorf("provider %s answered %s: %w", provider, code, app.ErrProviderUnavailable)
}

// providerFrom translates a Describe answer.
func providerFrom(d *modelsv1.DescribeResponse) (modelrouter.Provider, error) {
	p := modelrouter.Provider{RetentionDays: d.GetRetention().GetDays(), MayTrain: d.GetRetention().GetMayTrain()}
	switch d.GetLocation() {
	case modelsv1.Location_LOCATION_OWNER_HARDWARE:
		p.Location = modelrouter.LocationOwnerHardware
	case modelsv1.Location_LOCATION_THIRD_PARTY_CLOUD, modelsv1.Location_LOCATION_UNSPECIFIED:
		// A provider that does not say where its models run is treated as
		// the third-party cloud, the more cautious reading.
		p.Location = modelrouter.LocationThirdPartyCloud
	}
	return p, nil
}

// errBadEntry reports a model entry the plugin listed that the core can't
// use.
var errBadEntry = errors.New("invalid model entry")

// entryFrom translates a model entry, with its prices as exact decimals.
func entryFrom(m *modelsv1.ModelEntry) (modelrouter.Entry, error) {
	e := modelrouter.Entry{
		ID: m.GetId(), ContextTokens: m.GetContextTokens(), MaxOutputTokens: m.GetMaxOutputTokens(),
		ToolCalling: m.GetToolCalling(), StructuredOut: m.GetStructuredOutput(),
	}
	if e.ID == "" {
		return modelrouter.Entry{}, fmt.Errorf("%w: an entry without an id", errBadEntry)
	}
	if m.GetInputPrice() == "" && m.GetOutputPrice() == "" {
		return e, nil
	}
	var err error
	if e.InputPrice, err = kernel.ParseDecimal(m.GetInputPrice()); err != nil {
		return modelrouter.Entry{}, fmt.Errorf("%w: an input price that is not a decimal", errBadEntry)
	}
	if e.OutputPrice, err = kernel.ParseDecimal(m.GetOutputPrice()); err != nil {
		return modelrouter.Entry{}, fmt.Errorf("%w: an output price that is not a decimal", errBadEntry)
	}
	if e.PriceCurrency, err = kernel.ParseCurrency(m.GetPriceCurrency()); err != nil {
		return modelrouter.Entry{}, fmt.Errorf("%w: a price currency that is not an ISO 4217 code", errBadEntry)
	}
	e.HasPrice = true
	return e, nil
}

// requestTo translates a request for model.
func requestTo(model string, r *modelrouter.Request) (*modelsv1.GenerateRequest, error) {
	out := &modelsv1.GenerateRequest{Model: model, ResponseSchemaJson: r.ResponseSchemaJSON, MaxOutputTokens: r.MaxOutputTokens}
	for _, m := range r.Messages {
		role, err := roleTo(m.Role)
		if err != nil {
			return nil, err
		}
		pm := &modelsv1.Message{Role: role, Content: m.Content, ToolCallId: m.ToolCallID}
		for _, tc := range m.ToolCalls {
			pm.ToolCalls = append(pm.ToolCalls, toolCallTo(tc))
		}
		out.Messages = append(out.Messages, pm)
	}
	for _, t := range r.Tools {
		out.Tools = append(out.Tools, &modelsv1.Tool{Name: t.Name, Description: t.Description, ParametersSchemaJson: t.ParametersSchemaJSON})
	}
	return out, nil
}

func toolCallTo(tc modelrouter.ToolCall) *modelsv1.ToolCall {
	return &modelsv1.ToolCall{Id: tc.ID, Name: tc.Name, ArgumentsJson: tc.ArgumentsJSON}
}

// roleTo translates a message role.
func roleTo(r modelrouter.MessageRole) (modelsv1.MessageRole, error) {
	switch r {
	case modelrouter.MessageSystem:
		return modelsv1.MessageRole_MESSAGE_ROLE_SYSTEM, nil
	case modelrouter.MessageUser:
		return modelsv1.MessageRole_MESSAGE_ROLE_USER, nil
	case modelrouter.MessageAssistant:
		return modelsv1.MessageRole_MESSAGE_ROLE_ASSISTANT, nil
	case modelrouter.MessageTool:
		return modelsv1.MessageRole_MESSAGE_ROLE_TOOL, nil
	}
	return modelsv1.MessageRole_MESSAGE_ROLE_UNSPECIFIED, modelrouter.ErrInvalidRequest
}

// resultFrom translates a Generate answer.
func resultFrom(r *modelsv1.GenerateResponse) modelrouter.Result {
	out := modelrouter.Result{
		Content: r.GetContent(),
		Usage:   modelrouter.Usage{InputTokens: r.GetUsage().GetInputTokens(), OutputTokens: r.GetUsage().GetOutputTokens()},
	}
	switch r.GetFinishReason() {
	case modelsv1.FinishReason_FINISH_REASON_LENGTH:
		out.Finish = modelrouter.FinishLength
	case modelsv1.FinishReason_FINISH_REASON_TOOL_CALLS:
		out.Finish = modelrouter.FinishToolCalls
	case modelsv1.FinishReason_FINISH_REASON_STOP, modelsv1.FinishReason_FINISH_REASON_UNSPECIFIED:
		out.Finish = modelrouter.FinishStop
	}
	for _, tc := range r.GetToolCalls() {
		out.ToolCalls = append(out.ToolCalls, modelrouter.ToolCall{ID: tc.GetId(), Name: tc.GetName(), ArgumentsJSON: tc.GetArgumentsJson()})
	}
	return out
}
