package ui

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// The metadata of a refusal of GenerateAsk by a guard (D-208), as the core's
// AskModelService documents it (proto/openfief/models/v1/ask.proto): which
// guard, and, for the budget, when the day ends. The UI imports no OpenFief
// package but the SDK and its generated clients, so the names are spelled here.
const (
	refusalHeader  = "Openfief-Refusal"
	resetsAtHeader = "Openfief-Resets-At"
	refusalBudget  = "budget"
	refusalBusy    = "busy"
)

// admitKind is the core's decision on a question about to go to a model
// (AdmitAskQuestion).
type admitKind int

// The decisions. The zero decision allows nothing, so a core that answers none
// is not taken for one that allows.
const (
	admitUnknown admitKind = iota
	admitAllowed
	admitBudgetSpent
	admitSessionLimit
	admitAddressLimit
)

// askAdmission is a decision with, for a refusal, when its reason ends; the
// zero time when the core named none.
type askAdmission struct {
	Kind     admitKind
	ResetsAt time.Time
}

// clientAddress returns the browser's address as the core handed it to the UI
// (openfief.ClientIPHeader), "" when it could not tell. Only the core can reach
// the UI (requireCore), and it drops what a browser sent in that header, so it
// is believed; it is used to count questions and never to decide who may read.
func clientAddress(r *http.Request) string {
	return r.Header.Get(openfief.ClientIPHeader)
}

// humanWait says how long it is until a limit ends, as the Ask box tells a
// person: "less than a minute", "about 5 minutes", "about 3 hours". It rounds
// up to the minute and speaks of hours from two on, so the person never
// returns before it has ended.
func humanWait(d time.Duration) string {
	switch minutes := int((d + time.Minute - 1) / time.Minute); {
	case d <= 0, minutes <= 1:
		return "less than a minute"
	case time.Duration(minutes)*time.Minute < 2*time.Hour:
		return fmt.Sprintf("about %d minutes", minutes)
	default:
		return fmt.Sprintf("about %d hours", int((d+time.Hour-1)/time.Hour))
	}
}

// resetsPhrase is "in about N minutes" for a limit that ends at at, or "later"
// when the core named no time or it has passed.
func resetsPhrase(at, now time.Time) string {
	if at.IsZero() || !at.After(now) {
		return "later"
	}
	return "in " + humanWait(at.Sub(now))
}

// budgetSpentAnswer is the answer when the day's budget for the model is used up
// (D-208): it says so and when it starts over, and keeps the commands.
func budgetSpentAnswer(question string, resetsAt, now time.Time) answerCard {
	return modelCard(question, "The day's budget for the model is used up",
		"The installation sets a limit on what the model may cost in a day, and today's is spent. It starts over "+
			resetsPhrase(resetsAt, now)+". Until then the commands still answer from your own record.")
}

// sessionLimitAnswer is the answer when the session has asked the model as many
// questions as it may (D-208).
func sessionLimitAnswer(question string, resetsAt, now time.Time) answerCard {
	return modelCard(question, "You've asked the model enough for now",
		"This session has used the questions it may ask the model in the hour or the day. You can ask it again "+
			resetsPhrase(resetsAt, now)+". The commands still answer from your own record.")
}

// addressLimitAnswer is the answer when the browser's address has asked the model
// as many questions as it may, whichever sessions they came from (D-208).
func addressLimitAnswer(question string, resetsAt, now time.Time) answerCard {
	return modelCard(question, "Too many questions from your network",
		"Your network address has used the questions it may ask the model in the hour or the day, from every session on it. "+
			"You can ask it again "+resetsPhrase(resetsAt, now)+". The commands still answer from your own record.")
}

// busyAnswer is the answer when the model's calls at once were all taken for the
// whole wait (D-208).
func busyAnswer(question string) answerCard {
	return modelCard(question, "The model is busy",
		"Other questions are using it right now. Try again in a few seconds; the commands still answer from your own record.")
}

// admissionAnswer returns the card of a refusal, or false for a decision that
// lets the question go.
func admissionAnswer(question string, adm askAdmission, now time.Time) (answerCard, bool) {
	switch adm.Kind {
	case admitBudgetSpent:
		return budgetSpentAnswer(question, adm.ResetsAt, now), true
	case admitSessionLimit:
		return sessionLimitAnswer(question, adm.ResetsAt, now), true
	case admitAddressLimit:
		return addressLimitAnswer(question, adm.ResetsAt, now), true
	case admitAllowed:
		return answerCard{}, false
	case admitUnknown:
	}
	// A core that gave no decision is not one that said yes: the box asks nothing.
	return modelCard(question, "The model couldn't be checked",
		"The box couldn't confirm that the model may answer now. Try again in a moment; the commands answer from your own record."), true
}

// refusedByAGuard returns the card for a call the core refused for a guard, from
// the error's metadata, or false for any other error.
func refusedByAGuard(question string, err error, now time.Time) (answerCard, bool) {
	ce, ok := errors.AsType[*connect.Error](err)
	if !ok || ce.Code() != connect.CodeResourceExhausted {
		return answerCard{}, false
	}
	switch ce.Meta().Get(refusalHeader) {
	case refusalBudget:
		// A reset time that doesn't parse is the zero time, which the card words as "later".
		resets, parseErr := time.Parse(time.RFC3339, ce.Meta().Get(resetsAtHeader))
		if parseErr != nil {
			resets = time.Time{}
		}
		return budgetSpentAnswer(question, resets, now), true
	case refusalBusy:
		return busyAnswer(question), true
	}
	return answerCard{}, false
}
