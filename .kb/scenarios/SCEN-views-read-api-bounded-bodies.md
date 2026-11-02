---
title: A compressed or oversized read API request is refused before its token is checked
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T19:45:12.154Z'
id: SCEN-views-read-api-bounded-bodies
type: scenario
---
Given a caller without a token, when it sends the read API a gzip body that would inflate to megabytes, over Connect, gRPC-Web or REST, then it gets 415 and nothing of the body is read; when it sends a body declared over the cap it gets 413; and when it streams a body of unknown length past 1 MiB, the read stops just past the cap and the answer is resource exhausted, never unauthenticated.