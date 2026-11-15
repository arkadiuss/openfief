package main

import (
	"crypto/sha256"
	"encoding/hex"
	"unicode/utf8"

	"github.com/arkadiuss/openfief/plugins/model-openai/openaichat"
)

// replyTextLimit is how many characters of a reply's plain text a round keeps.
const replyTextLimit = 500

// hashChars is how many hex characters of a message's SHA-256 a round keeps.
const hashChars = 8

// RoundReport is one call of a run to the model, as report.json traces it: what
// the bench sent and what the host reported back. It holds no text of what was
// sent, only each message's role, length and a short hash, so the non-public
// part of the prompt stays out of the report; the reply's own text is kept, cut
// to 500 characters.
type RoundReport struct {
	// Round counts the calls of the run from 1; Retry marks the second attempt
	// of a call that failed at the transport level.
	Round int  `json:"round"`
	Retry bool `json:"retry,omitempty"`
	// Roles are the roles of the messages sent, in order, and Messages their count.
	Roles    []string `json:"roles"`
	Messages int      `json:"messages"`
	// Sent describes each message sent, in the order of Roles.
	Sent []SentMessage `json:"sent"`
	// ToolsOffered is how many tools the request offered.
	ToolsOffered int `json:"toolsOffered"`
	// PromptTokens and CompletionTokens are what the host reported, and
	// CachedTokens how many of the prompt tokens it said it served from its
	// cache (0 when it said nothing).
	PromptTokens     int64 `json:"promptTokens"`
	CompletionTokens int64 `json:"completionTokens"`
	CachedTokens     int64 `json:"cachedTokens,omitempty"`
	// FinishReason is the reply's; ToolCalls the names of the tools it asked for.
	FinishReason string   `json:"finishReason,omitempty"`
	ToolCalls    []string `json:"toolCalls,omitempty"`
	// Text is the reply's plain text, cut to 500 characters.
	Text string `json:"text,omitempty"`
	// Answered is false for a call that failed, whose reply fields are empty.
	Answered bool `json:"answered"`
}

// SentMessage is a message of a request as the trace keeps it: its role, its
// length in characters, a short hash of its content and the names of the tool
// calls it holds.
type SentMessage struct {
	Role  string   `json:"role"`
	Chars int      `json:"chars"`
	Hash  string   `json:"hash"`
	Calls []string `json:"calls,omitempty"`
}

// sentRound returns the sent half of a round for req.
func sentRound(req *openaichat.Request, round int, retry bool) RoundReport {
	r := RoundReport{Round: round, Retry: retry, Messages: len(req.Messages), ToolsOffered: len(req.Tools)}
	for i := range req.Messages {
		m := &req.Messages[i]
		sum := sha256.Sum256([]byte(m.Content))
		sm := SentMessage{Role: m.Role, Chars: utf8.RuneCountInString(m.Content), Hash: hex.EncodeToString(sum[:])[:hashChars]}
		for _, c := range m.ToolCalls {
			sm.Calls = append(sm.Calls, c.Name)
		}
		r.Roles = append(r.Roles, m.Role)
		r.Sent = append(r.Sent, sm)
	}
	return r
}

// fillReply adds what the host reported of an answered call.
func (r *RoundReport) fillReply(reply *openaichat.Reply) {
	r.Answered = true
	r.PromptTokens, r.CompletionTokens, r.CachedTokens = reply.InputTokens, reply.OutputTokens, reply.CachedTokens
	r.FinishReason = reply.FinishReason
	for _, tc := range reply.ToolCalls {
		r.ToolCalls = append(r.ToolCalls, tc.Name)
	}
	r.Text = truncateChars(reply.Content, replyTextLimit)
}

// truncateChars cuts s to at most limit characters.
func truncateChars(s string, limit int) string {
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	return string([]rune(s)[:limit])
}
