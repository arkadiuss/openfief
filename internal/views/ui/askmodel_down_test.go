package ui

import (
	"net/http"
	"testing"

	"connectrpc.com/connect"
)

// TestAskBoxWithEveryModelDownStillAnswersItsCommands checks D-203 and D-207
// for a demo whose models are down (the router answers Unavailable: the primary
// and its fallback both failed): the fixed commands still answer themselves, at
// once, from the person's own record and without asking the router, and any
// other question gets a card that says the model didn't answer and lists the
// commands. It does so for the owner and for a read session, the demo's visitor.
func TestAskBoxWithEveryModelDownStillAnswersItsCommands(t *testing.T) {
	t.Parallel()
	for _, session := range []string{ownerSession, readSession} {
		t.Run(session, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			h.askModel.turnOn("T0", "T2")
			h.askModel.failWith = codeErr(connect.CodeUnavailable, nil)
			for _, q := range []string{"What needs me?", "allocation by currency", "value on 2026-03-31", "show synthetic brokerage"} {
				rec := h.asks(t, session, q)
				wantStatus(t, rec, http.StatusOK)
				wantBody(t, rec, "Answered from your data · no AI model")
				wantNoBody(t, rec, "didn&#39;t answer", "Answered by")
			}
			if n := h.askModel.generated(); n != 0 {
				t.Fatalf("the router was asked %d times for commands, want none", n)
			}
			rec := h.asks(t, session, "how diversified am I?")
			wantStatus(t, rec, http.StatusOK)
			wantBody(t, rec, "The model didn&#39;t answer", "Every model on the route was down or too slow",
				"The model gave no answer", `class="ask-help"`, "What needs me", "Allocation by currency")
			wantNoBody(t, rec, "Answered by", `class="used-card"`)
		})
	}
}
