package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
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
//     before the next message that is not a tool's;
//   - every tool has a name of 1 to 64 letters, digits, underscores or dashes,
//     used once, and parameters that are a JSON object; a response schema is a
//     JSON object. These are what hosts refuse most often with a 400 that no
//     other host would answer differently: a 4xx that no entry of a run accepted
//     is the bench's fault (rejections.go) and means this check misses something.
//
// A request that fails is never sent: the run it belongs to is void, the bench's
// own fault and no host's.
func validateRequest(req *openaichat.Request) error {
	if err := checkTools(req); err != nil {
		return err
	}
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

// maxToolName is the longest name a chat API takes for a tool.
const maxToolName = 64

// toolNamePattern is a tool name a chat API takes: 1 to maxToolName letters,
// digits, underscores or dashes.
var toolNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

// validToolName reports whether name is one a chat API takes.
func validToolName(name string) bool { return toolNamePattern.MatchString(name) }

// isJSONObject reports whether raw is a JSON object.
func isJSONObject(raw json.RawMessage) bool {
	var v map[string]json.RawMessage
	return json.Unmarshal(raw, &v) == nil && v != nil
}

// checkTools checks the tools and the response schema of the request.
func checkTools(req *openaichat.Request) error {
	seen := map[string]bool{}
	for i := range req.Tools {
		t := &req.Tools[i]
		if !validToolName(t.Name) || seen[t.Name] {
			return fmt.Errorf("%w: tool %d: a name that is not 1 to %d letters, digits, underscores or dashes, or is used twice", errBadConversation, i, maxToolName)
		}
		seen[t.Name] = true
		if len(t.Parameters) > 0 && !isJSONObject(t.Parameters) {
			return fmt.Errorf("%w: tool %d: parameters that are not a JSON object", errBadConversation, i)
		}
	}
	if len(req.ResponseSchema) > 0 && !isJSONObject(req.ResponseSchema) {
		return fmt.Errorf("%w: a response schema that is not a JSON object", errBadConversation)
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
