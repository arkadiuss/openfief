---
title: A failed call's entry in the report names its error class, and a bad reply is counted in its own line
status: active
origin:
  kind: agent
  ref: 'Brief on run 16-1: three calls failed as "failed" with no cause in report.json (D-207, D-217), 2026-11-15'
  recorded_at: '2026-11-15T18:32:35.257Z'
id: SCEN-askbench-failed-call-has-a-class
type: scenario
---
Given a model that calls a tool, says nothing after its result and answers the final round with a 2xx reply that is no chat completion, then each run's report entry holds one call error with the label and class bad reply and the fault host, the call and its retry count as transport failures, and the report carries no provider text; given a reply that stopped at the length limit with no text or holds harmony control tokens, then it is a degenerate generation, the host's fault, retried once; given a conversation that breaks the chat API's rules (an assistant message with neither text nor tool calls, a tool call never answered, a tool message that answers none), then nothing is sent, the call fails as an invalid request, the bench's fault, and the run is void: it is marked void (bench fault) in summary.md and on the results page and has no place in the history; and given an error that is none of the chat client's, then its class is local error and it is the bench's fault, so that every failed call has a class and one bucket.