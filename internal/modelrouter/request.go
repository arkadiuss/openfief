package modelrouter

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrInvalidRequest reports a request no model could be asked: no messages,
// one that ends with the assistant's own, a tool without a name or an
// unknown message role. The error never quotes the request's text.
var ErrInvalidRequest = errors.New("invalid model request")

// MessageRole is the role of a message in a conversation.
type MessageRole string

// The message roles.
const (
	MessageSystem    MessageRole = "system"
	MessageUser      MessageRole = "user"
	MessageAssistant MessageRole = "assistant"
	// MessageTool carries the result of a tool call the model asked for.
	MessageTool MessageRole = "tool"
)

// ToolCall is a call of a tool the model asked for.
type ToolCall struct {
	ID   string
	Name string
	// ArgumentsJSON is the arguments as a JSON object text.
	ArgumentsJSON string
}

// Message is one message of a conversation, with the data tier of what it
// holds (D-206): the system prompt of a role is usually T2, a tool's result
// carries the tier of the data the tool read, and an answer the model gave
// carries the tier of the data it was given.
type Message struct {
	Role       MessageRole
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
	Tier       Tier
}

// Tool is a function the model may call. A tool's name, description and
// schema describe the tool and hold no portfolio data, so they carry no
// tier.
type Tool struct {
	Name        string
	Description string
	// ParametersSchemaJSON is a JSON Schema of the arguments.
	ParametersSchemaJSON string
}

// Request is what a role asks a model.
type Request struct {
	Messages []Message
	Tools    []Tool
	// ResponseSchemaJSON, when set, asks for an answer that is a JSON text
	// of this JSON Schema (structured output).
	ResponseSchemaJSON string
	// MaxOutputTokens bounds the answer; zero means the entry's limit.
	MaxOutputTokens int64
}

// Validate returns nil when r could be asked of a model, otherwise an error
// wrapping ErrInvalidRequest.
func (r *Request) Validate() error {
	if len(r.Messages) == 0 {
		return fmt.Errorf("%w: no messages", ErrInvalidRequest)
	}
	if r.Messages[len(r.Messages)-1].Role == MessageAssistant {
		return fmt.Errorf("%w: the last message is the assistant's", ErrInvalidRequest)
	}
	for _, m := range r.Messages {
		switch m.Role {
		case MessageSystem, MessageUser, MessageAssistant, MessageTool:
		default:
			return fmt.Errorf("%w: unknown message role", ErrInvalidRequest)
		}
	}
	for _, t := range r.Tools {
		if t.Name == "" {
			return fmt.Errorf("%w: a tool without a name", ErrInvalidRequest)
		}
	}
	if r.MaxOutputTokens < 0 {
		return fmt.Errorf("%w: negative output limit", ErrInvalidRequest)
	}
	return nil
}

// Tiers returns the distinct data tiers of r's messages, or an error
// wrapping ErrUntagged when a message carries none (D-206).
func (r *Request) Tiers() (TierSet, error) {
	tags := make([]Tier, 0, len(r.Messages))
	for _, m := range r.Messages {
		tags = append(tags, m.Tier)
	}
	return NewTierSet(tags)
}

// Needs returns what a model must support to serve r.
func (r *Request) Needs() Needs {
	return Needs{Tools: len(r.Tools) > 0, StructuredOutput: r.ResponseSchemaJSON != ""}
}

// Needs is what a request asks of a model entry.
type Needs struct {
	Tools            bool
	StructuredOutput bool
}

// Entry is a model entry of a provider instance (D-205): what the plugin
// lists for one model.
type Entry struct {
	// ID is the model's id at its provider.
	ID string
	// HasPrice is false for an entry with no price, such as a local model;
	// otherwise InputPrice and OutputPrice are the price per million tokens
	// in PriceCurrency.
	HasPrice      bool
	InputPrice    kernel.Decimal
	OutputPrice   kernel.Decimal
	PriceCurrency kernel.Currency
	// ContextTokens and MaxOutputTokens are the entry's limits, in tokens;
	// zero when unknown.
	ContextTokens   int64
	MaxOutputTokens int64
	ToolCalling     bool
	StructuredOut   bool
}

// Serves reports whether e supports everything n asks.
func (e *Entry) Serves(n Needs) bool {
	return (!n.Tools || e.ToolCalling) && (!n.StructuredOutput || e.StructuredOut)
}

// Usage is what one call used, as the provider reports it.
type Usage struct {
	InputTokens  int64
	OutputTokens int64
}

// FinishReason is why a model stopped.
type FinishReason string

// The finish reasons.
const (
	FinishStop      FinishReason = "stop"
	FinishLength    FinishReason = "length"
	FinishToolCalls FinishReason = "tool_calls"
)

// Result is a model's answer to a request.
type Result struct {
	// Model is the entry that answered.
	Model     Ref
	Content   string
	ToolCalls []ToolCall
	Finish    FinishReason
	Usage     Usage
}

// Location is where a provider instance's models run.
type Location string

// The locations.
const (
	// LocationOwnerHardware is hardware the owner runs.
	LocationOwnerHardware Location = "owner_hardware"
	// LocationThirdPartyCloud is a third party's cloud.
	LocationThirdPartyCloud Location = "third_party_cloud"
)

// Provider is what a model-provider plugin declares about itself (D-205):
// where its models run and its provider's retention terms.
type Provider struct {
	Location Location
	// RetentionDays is how long the provider keeps requests and answers;
	// zero for none.
	RetentionDays int32
	MayTrain      bool
}
