---
title: The Ghostfolio client logs in with the security token and keeps every request safe to repeat
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:33.783Z'
id: SCEN-views-ghostfolio-client
type: scenario
---
Given the owner's security token in its Secret, when the client logs in, then it exchanges the token for a bearer token at /api/v1/auth/anonymous; when it lists or bulk-deletes activities with no account, then it refuses without sending anything; when it deletes an activity Ghostfolio no longer has, then that counts as done; when Ghostfolio answers with an error status, then the error names the call and the status and never the response body.