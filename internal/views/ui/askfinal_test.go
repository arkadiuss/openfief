package ui

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// The final round and the retry send the whole conversation: the model must see
// its question, every tool call and result, and its own reply when it is asked
// again, or it answers as if it had never been asked.

func TestAPlainFirstReplyIsFollowedByTheWholeConversationAndTheInstruction(t *testing.T) {
	t.Parallel()
	const question = "Sell all my DTEC and buy bonds"
	const plain = "I can't change anything."
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.replies = []askReply{{Content: plain, Model: "p/m"}, sayKind("refuse", "I can't change anything.", nil, nil)}
	got, conv, err := f.askConv(question)
	if err != nil || got.Kind != answerKindRefuse || conv.rounds != 2 {
		t.Fatalf("answer = %+v, %v after %d calls; want the refusal after two", got, err, conv.rounds)
	}
	want := []askMessage{
		{Role: roleSystem, Content: askSystemPrompt(mustDate(t, askToday)), Tier: tierT2},
		{Role: roleUser, Content: question, Tier: tierT0},
		{Role: roleAssistant, Content: plain, Tier: tierT0},
		{Role: roleUser, Content: finalText, Tier: tierT2},
	}
	if diff := cmp.Diff(want, f.port.requests[1]); diff != "" {
		t.Errorf("messages of the second call (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]bool{false, true}, f.port.answers); diff != "" {
		t.Errorf("calls that asked for the answer (-want +got):\n%s", diff)
	}
}

func TestTheFinalRoundKeepsEveryToolCallAndResultAndTheModelsPlainReply(t *testing.T) {
	t.Parallel()
	const question = "How much is it worth?"
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.replies = []askReply{
		calling(askCall{ID: "1", Name: "value", Args: `{}`}),
		calling(askCall{ID: "2", Name: "allocation", Args: `{}`}),
		{Content: "It is 10.", Model: "p/m"},
		say("It is 10.", "c1"),
	}
	if _, conv, err := f.askConv(question); err != nil || conv.rounds != 4 {
		t.Fatalf("%v after %d calls, want four", err, conv.rounds)
	}
	before, final := f.port.requests[2], f.port.requests[3]
	if len(final) != len(before)+2 {
		t.Fatalf("the final round holds %d messages, want the %d before it and two more", len(final), len(before))
	}
	if diff := cmp.Diff(before, final[:len(before)]); diff != "" {
		t.Errorf("the final round lost or changed what came before (-want +got):\n%s", diff)
	}
	if final[1].Role != roleUser || final[1].Content != question {
		t.Errorf("the second message = %+v, want the question", final[1])
	}
	if got := len(toolMessages(final)); got != 2 {
		t.Errorf("the final round holds %d tool results, want 2", got)
	}
	wantTail := []askMessage{
		{Role: roleAssistant, Content: "It is 10.", Tier: tierT0},
		{Role: roleUser, Content: finalText, Tier: tierT2},
	}
	if diff := cmp.Diff(wantTail, final[len(final)-2:]); diff != "" {
		t.Errorf("the end of the final round (-want +got):\n%s", diff)
	}
}

func TestTheRetryAfterAnInvalidReplyKeepsTheWholeConversation(t *testing.T) {
	t.Parallel()
	const question = "Sell all my DTEC and buy bonds"
	const invalid = `{"kind":"nope"}`
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.replies = []askReply{
		{Content: "No.", Model: "p/m"},
		{Content: invalid, Model: "p/m"},
		sayKind("refuse", "No.", nil, nil),
	}
	if _, conv, err := f.askConv(question); err != nil || conv.rounds != 3 || conv.invalid != 1 {
		t.Fatalf("%v after %d calls and %d invalid, want three and one", err, conv.rounds, conv.invalid)
	}
	before, retry := f.port.requests[1], f.port.requests[2]
	if len(retry) != len(before)+2 {
		t.Fatalf("the retry holds %d messages, want the %d before it and two more", len(retry), len(before))
	}
	if diff := cmp.Diff(before, retry[:len(before)]); diff != "" {
		t.Errorf("the retry lost or changed what came before (-want +got):\n%s", diff)
	}
	if retry[1].Content != question || retry[0].Role != roleSystem {
		t.Errorf("the retry starts with %+v and %+v, want the system prompt and the question", retry[0].Role, retry[1])
	}
	wantTail := []askMessage{
		{Role: roleAssistant, Content: invalid, Tier: tierT0},
		{Role: roleUser, Content: retryText(errAnswerKind), Tier: tierT2},
	}
	if diff := cmp.Diff(wantTail, retry[len(retry)-2:]); diff != "" {
		t.Errorf("the end of the retry (-want +got):\n%s", diff)
	}
}

// The final round asks for the previous reply in the schema's form and nothing
// else, so a correct refusal in plain words stays a refusal (D-205, D-217).
func TestTheFinalRoundAsksForThePreviousReplyInTheSchemaWithoutChangingIt(t *testing.T) {
	t.Parallel()
	const instruction = "Put your previous reply into this schema without changing its meaning or adding anything."
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.replies = []askReply{{Content: "I can't change anything.", Model: "p/m"}, sayKind("refuse", "I can't change anything.", nil, nil)}
	if _, conv, err := f.askConv("Sell all my DTEC and buy bonds"); err != nil || conv.rounds != 2 {
		t.Fatalf("%v after %d calls, want two", err, conv.rounds)
	}
	final := f.port.requests[1]
	if last := final[len(final)-1]; last.Role != roleUser || last.Content != instruction {
		t.Errorf("the last message of the final round = %+v, want the user's %q", last, instruction)
	}
	if !strings.Contains(retryText(errAnswerKind), "without changing its meaning or adding anything") {
		t.Errorf("the retry = %q, want it to fix the form and keep the meaning", retryText(errAnswerKind))
	}
}

// A model that stops after a tool's result without a word has no reply to put
// into the schema: the final round asks for the answer from the cards, sends no
// empty assistant message and never says to restate (D-205, D-217).
func TestTheFinalRoundAfterAnEmptyReplyAsksForTheAnswerFromTheCards(t *testing.T) {
	t.Parallel()
	const question = "How much is it worth?"
	for name, empty := range map[string]string{"no text": "", "blank text": " \n\t "} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			f := newLoopFixture(t, tierT0, tierT2)
			f.port.replies = []askReply{
				calling(askCall{ID: "1", Name: "value", Args: `{}`}),
				{Content: empty, Model: "p/m"},
				say("It is 10.", "c1"),
			}
			got, conv, err := f.askConv(question)
			if err != nil || got.Kind != answerKindAnswer || conv.rounds != 3 {
				t.Fatalf("answer = %+v, %v after %d calls; want the answer after three", got, err, conv.rounds)
			}
			before, final := f.port.requests[1], f.port.requests[2]
			if len(final) != len(before)+1 {
				t.Fatalf("the final round holds %d messages, want the %d before it and the instruction", len(final), len(before))
			}
			if diff := cmp.Diff(before, final[:len(before)]); diff != "" {
				t.Errorf("the final round lost or changed what came before (-want +got):\n%s", diff)
			}
			checkEmptyReplyInstruction(t, final)
		})
	}
}

// checkEmptyReplyInstruction checks that the final round ends with the answer
// instruction, holds no empty assistant message, and asks to restate nothing.
func checkEmptyReplyInstruction(t *testing.T, final []askMessage) {
	t.Helper()
	want := askMessage{Role: roleUser, Content: "Answer the user's question in this schema, using only the cards above.", Tier: tierT2}
	if last := final[len(final)-1]; !cmp.Equal(last, want) {
		t.Errorf("the last message = %+v, want %+v", last, want)
	}
	for _, m := range final {
		if m.Role == roleAssistant && m.Content == "" && len(m.Calls) == 0 {
			t.Errorf("the final round holds an empty assistant message: %+v", m)
		}
	}
	if strings.Contains(strings.ToLower(want.Content), "previous reply") {
		t.Errorf("the instruction %q asks to restate a reply there is none of", want.Content)
	}
}
