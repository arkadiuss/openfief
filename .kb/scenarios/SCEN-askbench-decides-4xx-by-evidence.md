---
title: A 4xx is decided by whether another entry accepted the identical request, a 404 flags the pin and a 401 or 402 voids the run with exit 5
status: active
origin:
  kind: agent
  ref: 'Brief of the Ask bench 4xx and per-question reliability change: the design delegate''s decisions on D-207 and D-221, 2026-11-15'
  recorded_at: '2026-11-15T20:00:00.000Z'
id: SCEN-askbench-decides-4xx-by-evidence
type: scenario
---
Given two model@host entries asked the same questions, when one entry's host answers every request 400, 413 or 422 and the other accepts them, then the refusing host's calls are its fault by the rule accepted elsewhere, counted in its reliability, and the run is not void; given every entry refusing, or the only entry refusing, then the call is the bench's fault by the rule accepted nowhere, the run is void with no place in the history, and summary.md and stdout say the request check should have caught it, with the request's hash in report.json; given a 404, then it is the host's fault by the rule not found, counted against it, never retried, and summary.md and stdout say the pin may be stale; given a 401 or a 402, then the run is void for the credentials, counts against no host or model, has no place in the history, summary.md and stdout say so and the bench exits 5, a code distinct from 0 to 4; and no failed call is left with no fault.