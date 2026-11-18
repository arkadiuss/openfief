package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/arkadiuss/openfief/plugins/model-openai/openaichat"
)

// A 4xx that is no refused key, no credits and no missing route says that some
// request was refused, and the run's evidence says whose fault it is (D-207):
// when an identical request was accepted by another model@host entry of the same
// run, the host that refused it cannot serve it, and the refusal is the host's;
// when no entry accepted it, the request is wrong, the bench's fault, and the
// run is void. A request is identical when its body, without the members that
// differ by entry (the model id, the provider routing, the reasoning control and
// the usage accounting), has the same hash.

// requestShortHash is how many hex characters of a request's hash are kept.
const requestShortHash = 16

// routingMembers are the members of a request body that differ by model@host
// entry and say nothing of what the model is asked: the model's id, the provider
// routing, the reasoning control and the usage accounting.
func routingMembers() []string {
	return []string{"model", "provider", memberReasoning, "usage"}
}

// requestHash is a canonical hash of the request that client sends for req,
// without its routing members: the messages, the tools, the response format and
// the output limit. The same request to another host or model has the same hash.
func requestHash(client *openaichat.Client, req *openaichat.Request) string {
	body := client.RequestBody(req)
	for _, key := range routingMembers() {
		delete(body, key)
	}
	raw, err := json.Marshal(body) // a map's keys are written in order
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])[:requestShortHash]
}

// isEvidenceRejection reports whether a status is a 4xx that the run's evidence
// decides: any 4xx but the refused key, no credits, a missing route, a request
// that ran out of time and a 429, which have rules of their own.
func isEvidenceRejection(status int) bool {
	switch status {
	case http.StatusUnauthorized, http.StatusPaymentRequired, http.StatusNotFound, http.StatusRequestTimeout, http.StatusTooManyRequests:
		return false
	}
	return status >= http.StatusBadRequest && status < http.StatusInternalServerError
}

// resolveRejections decides the fault of every 4xx that waits for the evidence
// of the run, now that every question is asked: the host's when another entry
// (another model index) had an identical request accepted, the bench's when none
// did. It records the rule in each call error, counts a host's one as a failed
// call, and settles the runs again.
func resolveRejections(tasks []task, results []RunReport) {
	accepted := map[string]map[int]bool{}
	for i := range results {
		for _, h := range results[i].accepted {
			if accepted[h] == nil {
				accepted[h] = map[int]bool{}
			}
			accepted[h][tasks[i].model] = true
		}
	}
	for i := range results {
		run := &results[i]
		for j := range run.CallErrors {
			if ce := &run.CallErrors[j]; ce.Fault == faultUndecided {
				run.decide(ce, acceptedElsewhere(accepted[ce.RequestHash], tasks[i].model))
			}
		}
		run.settle()
	}
}

// decide gives a 4xx of the run the host's fault by the rule accepted elsewhere,
// counted as a failed call, or the bench's by the rule accepted nowhere.
func (r *RunReport) decide(ce *CallErrorReport, elsewhere bool) {
	if elsewhere {
		ce.Fault, ce.Rule = faultHost, ruleAcceptedElsewhere
		r.TransportFailures++
		return
	}
	ce.Fault, ce.Rule = faultBench, ruleAcceptedNowhere
}

// acceptedElsewhere reports whether some entry other than model accepted the
// request: entries is the set of the entries that did.
func acceptedElsewhere(entries map[int]bool, model int) bool {
	for m := range entries {
		if m != model {
			return true
		}
	}
	return false
}
