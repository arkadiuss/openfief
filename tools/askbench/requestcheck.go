package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/arkadiuss/openfief/plugins/model-openai/openaichat"
)

// The roles of a chat conversation.
const (
	chatSystem    = "system"
	chatUser      = "user"
	chatAssistant = "assistant"
	chatTool      = "tool"
)

// errBadConversation is a conversation that breaks the chat API's rules.
var errBadConversation = errors.New("bad conversation")

// breaks is the error of message i, which breaks a rule for the reason why.
func breaks(i int, why string) error {
	return fmt.Errorf("%w: message %d: %s", errBadConversation, i, why)
}

// validateRequest checks a request against the chat API's rules before the bench
// sends it, so a failed call is told from a request that was never well-formed:
//
//   - the messages are not empty, a system message comes only first, and the
//     last message is the user's or a tool's, never the assistant's;
//   - every message has a known role; a user message has text;
//   - an assistant message has text or tool calls, never neither, and each tool
//     call has an id and a name;
//   - a tool message answers a tool call of the assistant message before it
//     that was not answered yet, and carries that call's id;
//   - every tool call of an assistant message is answered by a tool message
//     before the next message that is not a tool's.
//
// A request that fails is never sent: the run it belongs to is void, the bench's
// own fault and no host's.
func validateRequest(req *openaichat.Request) error {
	if len(req.Messages) == 0 {
		return fmt.Errorf("%w: no messages", errBadConversation)
	}
	pending := map[string]bool{}
	for i := range req.Messages {
		m := &req.Messages[i]
		if len(pending) > 0 && m.Role != chatTool {
			return breaks(i, "a tool call is not answered before this message")
		}
		if err := checkMessage(i, m, pending); err != nil {
			return err
		}
	}
	if len(pending) > 0 {
		return fmt.Errorf("%w: a tool call is not answered", errBadConversation)
	}
	if last := req.Messages[len(req.Messages)-1].Role; last != chatUser && last != chatTool {
		return fmt.Errorf("%w: the conversation ends with a %s message", errBadConversation, last)
	}
	return nil
}

// checkMessage checks message i, adding the tool calls it makes to pending and
// removing the one it answers.
func checkMessage(i int, m *openaichat.Message, pending map[string]bool) error {
	switch m.Role {
	case chatSystem:
		if i != 0 {
			return breaks(i, "a system message that is not first")
		}
	case chatUser:
		if strings.TrimSpace(m.Content) == "" {
			return breaks(i, "a user message with no text")
		}
	case chatAssistant:
		return checkAssistant(i, m, pending)
	case chatTool:
		if !pending[m.ToolCallID] {
			return breaks(i, "a tool message that answers no open tool call")
		}
		delete(pending, m.ToolCallID)
	default:
		return breaks(i, "a role that is not a chat role")
	}
	return nil
}

// checkAssistant checks assistant message i and opens its tool calls.
func checkAssistant(i int, m *openaichat.Message, pending map[string]bool) error {
	if strings.TrimSpace(m.Content) == "" && len(m.ToolCalls) == 0 {
		return breaks(i, "an assistant message with neither text nor tool calls")
	}
	for _, c := range m.ToolCalls {
		if c.ID == "" || c.Name == "" || pending[c.ID] {
			return breaks(i, "a tool call with no id, no name or a repeated id")
		}
		pending[c.ID] = true
	}
	return nil
}
