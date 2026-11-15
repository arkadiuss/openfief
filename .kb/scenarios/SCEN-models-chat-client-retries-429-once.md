---
title: A 429 is retried once after Retry-After, capped at ten seconds, and a second 429 is an unavailable provider
status: active
origin:
  kind: agent
  ref: 'Design session decision: 429 handling, retried once honouring Retry-After (D-207), 2026-11-15'
  recorded_at: '2026-11-15T09:14:03.733Z'
id: SCEN-models-chat-client-retries-429-once
type: scenario
---
Given a provider that answers 429 with Retry-After in seconds, as an HTTP date, with neither or with garbage, when the client completes a request, then it waits the header's time (at most ten seconds, one second without a usable header) and sends the same request once more; given a second 429, then the error is ErrUnavailable with status 429; given any other status, then there is one request and no wait; given a wait that would pass the deadline, then there is no retry.
