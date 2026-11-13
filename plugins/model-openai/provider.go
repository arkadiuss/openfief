package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"connectrpc.com/connect"

	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
)

const (
	// keyMount is where the manifest mounts the instance's optional Secret
	// <instance>--api-key (D-89), and keyFile its key.
	keyMount = "/etc/openfief/model-provider"
	keyFile  = "api-key"
	// maxMessages is the most messages one request holds.
	maxMessages = 4096
)

// errNoKey reports a key file that exists but is empty or unreadable. A
// missing Secret is not an error: the provider needs no key, as a local
// Ollama does.
var errNoKey = errors.New("the api-key Secret is unreadable or empty")

// errInvalidRequest reports a request the plugin refuses before it calls the
// provider; it never quotes the request (GR-12).
var errInvalidRequest = errors.New("invalid request")

// provider is ModelProviderService over one provider instance (D-205): it
// answers Describe and ListModels from the config and Generate through the
// chat client. It holds no request or answer text beyond a call, and logs
// none (GR-12).
type provider struct {
	modelsv1connect.UnimplementedModelProviderServiceHandler

	cfg  config
	chat *chatClient
}

var _ modelsv1connect.ModelProviderServiceHandler = (*provider)(nil)

// status is the Connect error with code for err.
func status(code connect.Code, err error) *connect.Error { return connect.NewError(code, err) }

// Describe says the protocol, where the models run and the retention terms.
func (p *provider) Describe(context.Context, *connect.Request[modelsv1.DescribeRequest]) (*connect.Response[modelsv1.DescribeResponse], error) {
	return connect.NewResponse(&modelsv1.DescribeResponse{
		Protocol:  modelsv1.WireProtocol_WIRE_PROTOCOL_OPENAI_COMPATIBLE,
		Location:  p.cfg.location,
		Retention: p.cfg.retention,
	}), nil
}

// ListModels returns the configured entries.
func (p *provider) ListModels(context.Context, *connect.Request[modelsv1.ListModelsRequest]) (*connect.Response[modelsv1.ListModelsResponse], error) {
	return connect.NewResponse(&modelsv1.ListModelsResponse{Models: p.cfg.models}), nil
}

// Generate asks the provider's model for one answer.
func (p *provider) Generate(ctx context.Context, req *connect.Request[modelsv1.GenerateRequest]) (*connect.Response[modelsv1.GenerateResponse], error) {
	wire, err := p.wireRequest(req.Msg)
	if err != nil {
		return nil, status(connect.CodeInvalidArgument, err)
	}
	ctx, cancel := context.WithTimeout(ctx, p.cfg.timeout)
	defer cancel()
	reply, err := p.chat.complete(ctx, wire)
	switch {
	case errors.Is(err, errRejected), errors.Is(err, errBadReply):
		return nil, status(connect.CodeInvalidArgument, err)
	case err != nil:
		return nil, status(connect.CodeUnavailable, err)
	}
	return connect.NewResponse(responseFrom(reply)), nil
}

// wireRequest checks req against the entry it names and the contract's
// rules, and returns the chat request. Its errors never quote the request.
func (p *provider) wireRequest(req *modelsv1.GenerateRequest) (*chatRequest, error) {
	entry, ok := p.cfg.entry(req.GetModel())
	if !ok {
		return nil, fmt.Errorf("%w: the model is not one of the plugin's entries", errInvalidRequest)
	}
	if err := checkAgainstEntry(req, entry); err != nil {
		return nil, err
	}
	out := &chatRequest{Model: entry.GetId(), MaxTokens: req.GetMaxOutputTokens()}
	if out.MaxTokens == 0 {
		out.MaxTokens = entry.GetMaxOutputTokens()
	}
	for _, m := range req.GetMessages() {
		wm, err := wireMessage(m)
		if err != nil {
			return nil, err
		}
		out.Messages = append(out.Messages, wm)
	}
	for _, t := range req.GetTools() {
		tool := chatTool{Name: t.GetName(), Description: t.GetDescription()}
		if s := t.GetParametersSchemaJson(); s != "" {
			tool.Parameters = []byte(s)
		}
		out.Tools = append(out.Tools, tool)
	}
	if s := req.GetResponseSchemaJson(); s != "" {
		out.ResponseSchema = []byte(s)
	}
	return out, nil
}

// checkAgainstEntry refuses a request the contract or the entry's flags and
// limits do not allow: no messages, too many, or one ending with the
// assistant's; a tool without a name, or tools or a schema for an entry
// without those flags; an output limit that is negative or above the
// entry's.
func checkAgainstEntry(req *modelsv1.GenerateRequest, entry *modelsv1.ModelEntry) error {
	msgs := req.GetMessages()
	switch {
	case len(msgs) == 0 || len(msgs) > maxMessages:
		return fmt.Errorf("%w: no messages, or too many", errInvalidRequest)
	case msgs[len(msgs)-1].GetRole() == modelsv1.MessageRole_MESSAGE_ROLE_ASSISTANT:
		return fmt.Errorf("%w: the last message is the assistant's", errInvalidRequest)
	}
	if err := checkCapabilities(req, entry); err != nil {
		return err
	}
	if limit := entry.GetMaxOutputTokens(); req.GetMaxOutputTokens() < 0 || (limit > 0 && req.GetMaxOutputTokens() > limit) {
		return fmt.Errorf("%w: the output limit is negative or above the entry's", errInvalidRequest)
	}
	return nil
}

// checkCapabilities refuses tools without a name, and tools or a response
// schema for an entry without those flags.
func checkCapabilities(req *modelsv1.GenerateRequest, entry *modelsv1.ModelEntry) error {
	if len(req.GetTools()) > 0 && !entry.GetToolCalling() {
		return fmt.Errorf("%w: the model's entry does not allow tools", errInvalidRequest)
	}
	if req.GetResponseSchemaJson() != "" && !entry.GetStructuredOutput() {
		return fmt.Errorf("%w: the model's entry does not allow structured output", errInvalidRequest)
	}
	for _, t := range req.GetTools() {
		if t.GetName() == "" {
			return fmt.Errorf("%w: a tool without a name", errInvalidRequest)
		}
	}
	return nil
}

// wireMessage translates one message.
func wireMessage(m *modelsv1.Message) (chatMessage, error) {
	roles := map[modelsv1.MessageRole]string{
		modelsv1.MessageRole_MESSAGE_ROLE_SYSTEM:    "system",
		modelsv1.MessageRole_MESSAGE_ROLE_USER:      "user",
		modelsv1.MessageRole_MESSAGE_ROLE_ASSISTANT: "assistant",
		modelsv1.MessageRole_MESSAGE_ROLE_TOOL:      "tool",
	}
	role, ok := roles[m.GetRole()]
	if !ok {
		return chatMessage{}, fmt.Errorf("%w: a message of an unknown role", errInvalidRequest)
	}
	out := chatMessage{Role: role, Content: m.GetContent(), ToolCallID: m.GetToolCallId()}
	for _, tc := range m.GetToolCalls() {
		out.ToolCalls = append(out.ToolCalls, chatToolCall{ID: tc.GetId(), Name: tc.GetName(), Arguments: tc.GetArgumentsJson()})
	}
	return out, nil
}

// responseFrom translates a reply.
func responseFrom(reply *chatReply) *modelsv1.GenerateResponse {
	out := &modelsv1.GenerateResponse{
		Content: reply.Content,
		Usage:   &modelsv1.Usage{InputTokens: reply.InputTokens, OutputTokens: reply.OutputTokens},
	}
	for _, tc := range reply.ToolCalls {
		out.ToolCalls = append(out.ToolCalls, &modelsv1.ToolCall{Id: tc.ID, Name: tc.Name, ArgumentsJson: tc.Arguments})
	}
	finish := map[string]modelsv1.FinishReason{
		"stop":       modelsv1.FinishReason_FINISH_REASON_STOP,
		"length":     modelsv1.FinishReason_FINISH_REASON_LENGTH,
		"tool_calls": modelsv1.FinishReason_FINISH_REASON_TOOL_CALLS,
	}
	out.FinishReason = finish[reply.FinishReason] // an unknown reason is unspecified
	return out
}

// keyFromMount returns the function that reads the instance's API key from
// the Secret mounted at dir on every call, through an os.Root so the read
// can't leave dir. A missing Secret reads as no key; an empty one is an
// error. The key is never part of an error.
func keyFromMount(dir string) func() (string, error) {
	return func() (string, error) {
		root, err := os.OpenRoot(dir)
		if errors.Is(err, fs.ErrNotExist) {
			return "", nil // the optional Secret is not mounted: the provider needs no key
		}
		if err != nil {
			return "", fmt.Errorf("%w: %s", errNoKey, dir)
		}
		b, readErr := root.ReadFile(keyFile)
		closeErr := root.Close()
		if errors.Is(readErr, fs.ErrNotExist) {
			return "", nil
		}
		v := strings.TrimSpace(string(b))
		if readErr != nil || closeErr != nil || v == "" {
			return "", fmt.Errorf("%w: %s", errNoKey, filepath.Join(dir, keyFile))
		}
		return v, nil
	}
}
