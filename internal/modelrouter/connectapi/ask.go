package connectapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
)

// The most one call of the Ask box's conversation may hold: its tool loop is
// bounded to a few rounds of a few tool calls (D-203), so this is far above
// what an honest caller sends.
const (
	maxAskMessages = 64
	maxAskTools    = 16
	// maxAskSchemaBytes caps the response schema of a call: the Ask box's is a
	// few hundred bytes.
	maxAskSchemaBytes = 4096
)

// NewAsk returns the AskModelService's path and HTTP handler over ask, with
// callers in front of every call to admit view plugins only, logging to log.
func NewAsk(ask *app.Ask, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return modelsv1connect.NewAskModelServiceHandler(&askService{ask: ask, log: log},
		connect.WithInterceptors(callers), connect.WithReadMaxBytes(maxRequestBytes))
}

// askService implements modelsv1connect.AskModelServiceHandler.
type askService struct {
	ask *app.Ask
	log *slog.Logger
}

// The metadata a guard's refusal of GenerateAsk carries (D-208): which guard,
// and, for the budget, when the day ends. The Ask box reads them to say so.
const (
	// RefusalHeader names the guard that refused a call: RefusalBudget or
	// RefusalBusy.
	RefusalHeader = "Openfief-Refusal"
	// ResetsAtHeader is when the day's budget starts over, RFC 3339.
	ResetsAtHeader = "Openfief-Resets-At"
	// RefusalBudget: the day's budget is spent.
	RefusalBudget = "budget"
	// RefusalBusy: the Ask box's calls at once were all taken.
	RefusalBusy = "busy"
)

// AdmitAskQuestion decides on a question the Ask box is about to send a model.
func (s *askService) AdmitAskQuestion(ctx context.Context, req *connect.Request[modelsv1.AdmitAskQuestionRequest]) (*connect.Response[modelsv1.AdmitAskQuestionResponse], error) {
	got, err := s.ask.Admit(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetClientAddress())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	out := &modelsv1.AdmitAskQuestionResponse{}
	switch got.Decision {
	case app.AdmitAllowed:
		out.Admission = modelsv1.AskAdmission_ASK_ADMISSION_ALLOWED
	case app.AdmitBudgetSpent:
		out.Admission = modelsv1.AskAdmission_ASK_ADMISSION_BUDGET_SPENT
	case app.AdmitSessionLimit:
		out.Admission = modelsv1.AskAdmission_ASK_ADMISSION_SESSION_LIMIT
	case app.AdmitAddressLimit:
		out.Admission = modelsv1.AskAdmission_ASK_ADMISSION_ADDRESS_LIMIT
	}
	if !got.ResetsAt.IsZero() {
		out.ResetsAt = timestamppb.New(got.ResetsAt)
	}
	return connect.NewResponse(out), nil
}

// GetAskModel says whether the Ask box has a model.
func (s *askService) GetAskModel(ctx context.Context, req *connect.Request[modelsv1.GetAskModelRequest]) (*connect.Response[modelsv1.GetAskModelResponse], error) {
	ok, err := s.ask.Available(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	opinions, err := s.ask.Opinions(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	out := &modelsv1.GetAskModelResponse{Available: ok}
	switch opinions {
	case modelrouter.OpinionsOn:
		out.Opinions = modelsv1.AskOpinions_ASK_OPINIONS_ON
	case modelrouter.OpinionsRouteIneligible:
		out.Opinions = modelsv1.AskOpinions_ASK_OPINIONS_ROUTE_INELIGIBLE
	case modelrouter.OpinionsOff:
	}
	return connect.NewResponse(out), nil
}

// CheckAskTiers says whether a model of the Ask box's route may receive
// every one of the given tiers together.
func (s *askService) CheckAskTiers(ctx context.Context, req *connect.Request[modelsv1.CheckAskTiersRequest]) (*connect.Response[modelsv1.CheckAskTiersResponse], error) {
	ok, err := s.ask.Permits(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetTiers())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&modelsv1.CheckAskTiersResponse{Allowed: ok}), nil
}

// GenerateAsk asks the Ask box's models for the next message.
func (s *askService) GenerateAsk(ctx context.Context, req *connect.Request[modelsv1.GenerateAskRequest]) (*connect.Response[modelsv1.GenerateAskResponse], error) {
	in, err := askRequestFrom(req.Msg)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	res, err := s.ask.Generate(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), in)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(askResponseTo(&res)), nil
}

// errBadAsk reports a request of the Ask box's contract that breaks it. It
// quotes nothing of the request (GR-12).
var errBadAsk = errors.New("invalid request")

// askRequestFrom translates the messages and tools of the Ask box's
// conversation. A message without a tier is left without one: the router
// refuses it and records the refusal.
func askRequestFrom(m *modelsv1.GenerateAskRequest) (*modelrouter.Request, error) {
	if len(m.GetMessages()) > maxAskMessages || len(m.GetTools()) > maxAskTools {
		return nil, fmt.Errorf("%w: too many messages or tools", errBadAsk)
	}
	out := &modelrouter.Request{MaxOutputTokens: m.GetMaxOutputTokens()}
	if schema := m.GetResponseSchemaJson(); schema != "" {
		if len(schema) > maxAskSchemaBytes || !json.Valid([]byte(schema)) {
			return nil, fmt.Errorf("%w: a response schema that is too long or not JSON", errBadAsk)
		}
		// The Ask box's structured answer: a model whose entry has no
		// structured output is asked for a JSON object instead, which the Ask
		// box checks against the schema.
		out.ResponseSchemaJSON, out.JSONModeFallback = schema, true
	}
	for _, am := range m.GetMessages() {
		msg, err := askMessageFrom(am)
		if err != nil {
			return nil, err
		}
		out.Messages = append(out.Messages, msg)
	}
	for _, t := range m.GetTools() {
		out.Tools = append(out.Tools, modelrouter.Tool{Name: t.GetName(), Description: t.GetDescription(), ParametersSchemaJSON: t.GetParametersSchemaJson()})
	}
	return out, nil
}

// askMessageFrom translates one message with its tier.
func askMessageFrom(am *modelsv1.AskMessage) (modelrouter.Message, error) {
	pm := am.GetMessage()
	role, err := messageRoleFrom(pm.GetRole())
	if err != nil {
		return modelrouter.Message{}, err
	}
	out := modelrouter.Message{Role: role, Content: pm.GetContent(), ToolCallID: pm.GetToolCallId()}
	if tier := am.GetDataTier(); tier != "" {
		if out.DataTier, err = modelrouter.ParseMessageTier(tier); err != nil {
			return modelrouter.Message{}, fmt.Errorf("%w: a data tier that is none of the four, or instructions", errBadAsk)
		}
	}
	for _, tc := range pm.GetToolCalls() {
		out.ToolCalls = append(out.ToolCalls, modelrouter.ToolCall{ID: tc.GetId(), Name: tc.GetName(), ArgumentsJSON: tc.GetArgumentsJson()})
	}
	return out, nil
}

// messageRoleFrom translates a message role.
func messageRoleFrom(r modelsv1.MessageRole) (modelrouter.MessageRole, error) {
	switch r {
	case modelsv1.MessageRole_MESSAGE_ROLE_SYSTEM:
		return modelrouter.MessageSystem, nil
	case modelsv1.MessageRole_MESSAGE_ROLE_USER:
		return modelrouter.MessageUser, nil
	case modelsv1.MessageRole_MESSAGE_ROLE_ASSISTANT:
		return modelrouter.MessageAssistant, nil
	case modelsv1.MessageRole_MESSAGE_ROLE_TOOL:
		return modelrouter.MessageTool, nil
	case modelsv1.MessageRole_MESSAGE_ROLE_UNSPECIFIED:
	}
	return "", fmt.Errorf("%w: a message role that is none of the four", errBadAsk)
}

// askResponseTo translates the model's answer. Only the output tokens go on,
// so the Ask box can stop a question at its token limit (D-208); the budget
// and the cost stay in the core.
func askResponseTo(r *modelrouter.Result) *modelsv1.GenerateAskResponse {
	out := &modelsv1.GenerateAskResponse{Content: r.Content, Model: r.Model.String(), OutputTokens: r.Usage.OutputTokens}
	switch r.Finish {
	case modelrouter.FinishLength:
		out.FinishReason = modelsv1.FinishReason_FINISH_REASON_LENGTH
	case modelrouter.FinishToolCalls:
		out.FinishReason = modelsv1.FinishReason_FINISH_REASON_TOOL_CALLS
	case modelrouter.FinishStop:
		out.FinishReason = modelsv1.FinishReason_FINISH_REASON_STOP
	}
	for _, tc := range r.ToolCalls {
		out.ToolCalls = append(out.ToolCalls, &modelsv1.ToolCall{Id: tc.ID, Name: tc.Name, ArgumentsJson: tc.ArgumentsJSON})
	}
	return out
}

// fail maps err to its Connect error, by the error's kind and never its text.
// Neither the question nor the answer reaches a log line (GR-12): an
// unexpected failure is logged by its safe text only.
func (s *askService) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrInvalidInput), errors.Is(err, errBadAsk), errors.Is(err, modelrouter.ErrUntagged), errors.Is(err, modelrouter.ErrInvalidRequest),
		errors.Is(err, app.ErrProviderRejected):
		return connect.NewError(connect.CodeInvalidArgument, errBadAsk)
	case errors.Is(err, modelrouter.ErrDataTierDenied), errors.Is(err, modelrouter.ErrNoCapableModel):
		return connect.NewError(connect.CodePermissionDenied, modelrouter.ErrDataTierDenied)
	case errors.Is(err, modelrouter.ErrSwitchOff), errors.Is(err, modelrouter.ErrFeatureOff), errors.Is(err, modelrouter.ErrNoRoute):
		return connect.NewError(connect.CodeFailedPrecondition, modelrouter.ErrSwitchOff)
	case errors.Is(err, modelrouter.ErrBudgetSpent):
		out := connect.NewError(connect.CodeResourceExhausted, modelrouter.ErrBudgetSpent)
		out.Meta().Set(RefusalHeader, RefusalBudget)
		if spent, ok := errors.AsType[*modelrouter.BudgetSpentError](err); ok {
			out.Meta().Set(ResetsAtHeader, spent.ResetsAt.UTC().Format(time.RFC3339))
		}
		return out
	case errors.Is(err, app.ErrBusy):
		out := connect.NewError(connect.CodeResourceExhausted, app.ErrBusy)
		out.Meta().Set(RefusalHeader, RefusalBusy)
		return out
	case errors.Is(err, app.ErrNoAnswer), errors.Is(err, app.ErrProviderUnavailable), errors.Is(err, context.DeadlineExceeded):
		return connect.NewError(connect.CodeUnavailable, app.ErrNoAnswer)
	case errors.Is(err, context.Canceled):
		return connect.NewError(connect.CodeCanceled, context.Canceled)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "ask model call failed", slog.String("error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded)))
	return connect.NewError(connect.CodeInternal, errInternal)
}
