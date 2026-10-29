---
title: The read API refuses a call without a script token
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T18:01:07.343Z'
id: SCEN-e2e-read-api-needs-token
type: scenario
---
Given the core ready, when the client pod calls GET /v1/accounts without an Authorization header, then the core answers 401 and returns no data.