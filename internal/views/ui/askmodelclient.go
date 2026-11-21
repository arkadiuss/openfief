package ui

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"

	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
)

// errAskMessage reports a message of the conversation this client can't send. It
// quotes nothing of it (GR-12).
var errAskMessage = errors.New("a message of the Ask box's conversation has no valid role")

// askModelClient is askModelPort over the core's AskModelService: the plugin
// presents its own token to be admitted, and the asking person's origin and
// session in each request, which the core checks. The questions and answers it
// carries are the person's data; it keeps and logs none (GR-12).
type askModelClient struct {
	c modelsv1connect.AskModelServiceClient
}

var _ askModelPort = askModelClient{}

// available asks whether the Ask box has a model for sess.
func (m askModelClient) available(ctx context.Context, sess session) (bool, error) {
	resp, err := m.c.GetAskModel(ctx, connect.NewRequest(&modelsv1.GetAskModelRequest{Origin: sess.Origin, SessionToken: sess.Token}))
	if err != nil {
		return false, fmt.Errorf("GetAskModel: %w", err)
	}
	return resp.Msg.GetAvailable(), nil
}

// opinions asks whether the Ask box may give opinions to the person of sess.
func (m askModelClient) opinions(ctx context.Context, sess session) (opinionsState, error) {
	resp, err := m.c.GetAskModel(ctx, connect.NewRequest(&modelsv1.GetAskModelRequest{Origin: sess.Origin, SessionToken: sess.Token}))
	if err != nil {
		return opinionsOff, fmt.Errorf("GetAskModel: %w", err)
	}
	switch resp.Msg.GetOpinions() {
	case modelsv1.AskOpinions_ASK_OPINIONS_ON:
		return opinionsOn, nil
	case modelsv1.AskOpinions_ASK_OPINIONS_ROUTE_INELIGIBLE:
		return opinionsIneligible, nil
	case modelsv1.AskOpinions_ASK_OPINIONS_UNSPECIFIED:
	}
	return opinionsOff, nil
}

// permits asks whether a model of the route may receive every one of tiers.
func (m askModelClient) permits(ctx context.Context, sess session, tiers []string) (bool, error) {
	resp, err := m.c.CheckAskTiers(ctx, connect.NewRequest(&modelsv1.CheckAskTiersRequest{Origin: sess.Origin, SessionToken: sess.Token, Tiers: tiers}))
	if err != nil {
		return false, fmt.Errorf("CheckAskTiers: %w", err)
	}
	return resp.Msg.GetAllowed(), nil
}

// admit asks whether the core lets the question go to the model, with the
// browser's address when it is known.
func (m askModelClient) admit(ctx context.Context, sess session, address string) (askAdmission, error) {
	resp, err := m.c.AdmitAskQuestion(ctx, connect.NewRequest(&modelsv1.AdmitAskQuestionRequest{
		Origin: sess.Origin, SessionToken: sess.Token, ClientAddress: address,
	}))
	if err != nil {
		return askAdmission{}, fmt.Errorf("AdmitAskQuestion: %w", err)
	}
	out := askAdmission{}
	if t := resp.Msg.GetResetsAt(); t != nil {
		out.ResetsAt = t.AsTime()
	}
	switch resp.Msg.GetAdmission() {
	case modelsv1.AskAdmission_ASK_ADMISSION_ALLOWED:
		out.Kind = admitAllowed
	case modelsv1.AskAdmission_ASK_ADMISSION_BUDGET_SPENT:
		out.Kind = admitBudgetSpent
	case modelsv1.AskAdmission_ASK_ADMISSION_SESSION_LIMIT:
		out.Kind = admitSessionLimit
	case modelsv1.AskAdmission_ASK_ADMISSION_ADDRESS_LIMIT:
		out.Kind = admitAddressLimit
	case modelsv1.AskAdmission_ASK_ADMISSION_UNSPECIFIED:
	}
	return out, nil
}

// generate asks the route's models for the next message of the conversation.
func (m askModelClient) generate(ctx context.Context, sess session, msgs []askMessage, tools []askTool, maxTokens int64, schema string) (askReply, error) {
	req := &modelsv1.GenerateAskRequest{Origin: sess.Origin, SessionToken: sess.Token, MaxOutputTokens: maxTokens, ResponseSchemaJson: schema}
	for i := range msgs {
		am, err := askMessageTo(&msgs[i])
		if err != nil {
			return askReply{}, err
		}
		req.Messages = append(req.Messages, am)
	}
	for i := range tools {
		req.Tools = append(req.Tools, &modelsv1.Tool{Name: tools[i].Name, Description: tools[i].Description, ParametersSchemaJson: tools[i].Schema})
	}
	resp, err := m.c.GenerateAsk(ctx, connect.NewRequest(req))
	if err != nil {
		return askReply{}, fmt.Errorf("GenerateAsk: %w", err)
	}
	out := askReply{Content: resp.Msg.GetContent(), Model: resp.Msg.GetModel(), OutputTokens: resp.Msg.GetOutputTokens()}
	for _, tc := range resp.Msg.GetToolCalls() {
		out.Calls = append(out.Calls, askCall{ID: tc.GetId(), Name: tc.GetName(), Args: tc.GetArgumentsJson()})
	}
	return out, nil
}

// askMessageTo translates a message of the conversation.
func askMessageTo(m *askMessage) (*modelsv1.AskMessage, error) {
	roles := map[string]modelsv1.MessageRole{
		roleSystem: modelsv1.MessageRole_MESSAGE_ROLE_SYSTEM, roleUser: modelsv1.MessageRole_MESSAGE_ROLE_USER,
		roleAssistant: modelsv1.MessageRole_MESSAGE_ROLE_ASSISTANT, roleTool: modelsv1.MessageRole_MESSAGE_ROLE_TOOL,
	}
	role, ok := roles[m.Role]
	if !ok {
		return nil, errAskMessage
	}
	pm := &modelsv1.Message{Role: role, Content: m.Content, ToolCallId: m.CallID}
	for _, c := range m.Calls {
		pm.ToolCalls = append(pm.ToolCalls, &modelsv1.ToolCall{Id: c.ID, Name: c.Name, ArgumentsJson: c.Args})
	}
	return &modelsv1.AskMessage{Message: pm, DataTier: m.Tier}, nil
}
