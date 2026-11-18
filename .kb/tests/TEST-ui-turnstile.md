---
title: The Ask box checks a person is asking before the model, with a fake Cloudflare
status: active
sourceFile: internal/views/ui/turnstile_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-14T22:47:02.428Z'
id: TEST-ui-turnstile
type: test
---
Against a fake siteverify server: the verdict for good, forged, used, missing, too long, mismatched-host and unreadable-secret tokens and for Cloudflare answering an error, no JSON or refusing the secret; what Cloudflare is sent (secret, token, and the address when known); a used token refused; Cloudflare unreachable failing closed; nothing secret logged; the config's validation; and through the page: the widget only where a model answers and Turnstile is on, the policy naming Cloudflare's origin alone, a good token passing, a missing, forged or unverifiable one stopping the question before anything is sent or counted, a single-use token refused the second time, commands needing none, and the check coming before the core's admission. Each verdict's card: no token, a token Cloudflare refused (with its error codes, other words dropped), a token made for another host (naming both hosts) and an unavailable check, none of them logging or showing the token, the secret or the question; and the page's widget carrying the callbacks and the status line.
